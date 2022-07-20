//go:build windows

package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/windows"

	isync "github.com/Microsoft/go-winio/internal/sync"
)

//sys cancelIoEx(file syscall.Handle, o *syscall.Overlapped) (err error) = CancelIoEx
//sys setFileCompletionNotificationModes(h syscall.Handle, flags uint8) (err error) = SetFileCompletionNotificationModes
//sys wsaGetOverlappedResult(h syscall.Handle, o *syscall.Overlapped, bytes *uint32, wait bool, flags *uint32) (err error) = ws2_32.WSAGetOverlappedResult

const (
	skipCompletionPortOnSuccess = 1
	skipSetEventOnHandle        = 2
)

var (
	ErrFileClosed = errors.New("file has already been closed")
	ErrTimeout    = &timeoutError{}
)

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

// Win32File implements Reader, Writer, and Closer on a Win32 handle without blocking in a syscall.
// It takes ownership of this handle and will close it if it is garbage collected.
type Win32File struct {
	Handle        syscall.Handle
	ReadDeadline  deadlineHandler
	WriteDeadline deadlineHandler

	isSocket bool
	// wgLock should be locked when closing or adding new operations to prevent new operations
	// from starting while mutating file state
	wgLock sync.RWMutex
	// wg is the group for all pending IO operations
	wg      sync.WaitGroup
	closing isync.AtomicBool
}

var _ io.ReadWriteCloser = &Win32File{}

// MakeWin32File makes a new win32File from an existing file handle
func MakeWin32File(h syscall.Handle, socket bool) (*Win32File, error) {
	f := &Win32File{
		Handle:        h,
		ReadDeadline:  newDeadlineHandler(),
		WriteDeadline: newDeadlineHandler(),
	}
	if err := createFileIoCompletionPort(windows.Handle(h)); err != nil {
		return nil, err
	}
	if err := setFileCompletionNotificationModes(h, skipCompletionPortOnSuccess|skipSetEventOnHandle); err != nil {
		return nil, err
	}
	return f, nil
}

// closeHandle closes the resources associated with a Win32 handle
func (f *Win32File) closeHandle() {
	f.wgLock.Lock()
	// Atomically set that we are closing, releasing the resources only once.
	if !f.closing.Swap(true) {
		f.wgLock.Unlock()
		// cancel all IO and wait for it to complete
		cancelIoEx(f.Handle, nil) //nolint:errcheck
		f.wg.Wait()
		// at this point, no new IO can start
		syscall.Close(f.Handle)
		f.Handle = 0
	} else {
		f.wgLock.Unlock()
	}
}

// Close closes a win32File.
func (f *Win32File) Close() error {
	f.closeHandle()
	return nil
}

// IsClosed checks if the file has been closed
func (f *Win32File) IsClosed() bool {
	return f.closing.IsSet()
}

// PrepareIo prepares for a new IO operation.
// The caller must call [IoOperation.Close] when the IO is finished, prior to [Close()] returning.
func (f *Win32File) PrepareIo() (*IoOperation, error) {
	f.wgLock.RLock()
	defer f.wgLock.RUnlock()
	if f.closing.IsSet() {
		return nil, ErrFileClosed
	}
	c := newIoOperation(f)
	return c, nil
}

// AsyncIo processes the return value from ReadFile or WriteFile, blocking until
// the operation has actually completed.
func (f *Win32File) AsyncIo(c *IoOperation, d deadline, bytes uint32, err error) (int, error) {
	//nolint:errorlint
	if err != syscall.ERROR_IO_PENDING {
		return int(bytes), err
	}

	if f.closing.IsSet() {
		cancelIoEx(f.Handle, &c.O) //nolint:errcheck
	}

	var r ioResult
	select {
	case r = <-c.ch:
		err = r.err
		if err == syscall.ERROR_OPERATION_ABORTED {
			if f.closing.IsSet() {
				err = ErrFileClosed
			}
		} else if err != nil && f.isSocket {
			// err is from Win32. Query the overlapped structure to get the winsock error.
			var bytes, flags uint32
			err = wsaGetOverlappedResult(f.Handle, &c.O, &bytes, false, &flags)
		}
	case <-d.Done():
		cancelIoEx(f.Handle, &c.O) //nolint:errcheck
		r = <-c.ch
		err = r.err
		if err == syscall.ERROR_OPERATION_ABORTED {
			err = ErrTimeout
		}
	}

	// runtime.KeepAlive is needed, as c is passed via native
	// code to ioCompletionProcessor, c must remain alive
	// until the channel read is complete.
	// todo: (de)allocate *ioOperation via win32 heap functions, instead of needing to KeepAlive?
	runtime.KeepAlive(c)
	return int(r.bytes), err
}

// Read reads from a file handle.
func (f *Win32File) Read(b []byte) (int, error) {
	c, err := f.PrepareIo()
	if err != nil {
		return 0, err
	}
	defer c.Close()

	if f.ReadDeadline.timedout.IsSet() {
		return 0, ErrTimeout
	}

	var bytes uint32
	err = syscall.ReadFile(f.Handle, b, &bytes, &c.O)
	n, err := f.AsyncIo(c, &f.ReadDeadline, bytes, err)
	// todo: is this needed for reads, since should AsyncIo block until the data is read from b
	runtime.KeepAlive(b)

	// Handle EOF conditions.
	if err == nil && n == 0 && len(b) != 0 {
		return 0, io.EOF
	} else if err == syscall.ERROR_BROKEN_PIPE {
		return 0, io.EOF
	} else {
		return n, err
	}
}

// Write writes to a file handle.
func (f *Win32File) Write(b []byte) (int, error) {
	c, err := f.PrepareIo()
	if err != nil {
		return 0, err
	}
	defer c.Close()

	if f.WriteDeadline.timedout.IsSet() {
		return 0, ErrTimeout
	}

	var bytes uint32
	err = syscall.WriteFile(f.Handle, b, &bytes, &c.O)
	n, err := f.AsyncIo(c, &f.WriteDeadline, bytes, err)
	runtime.KeepAlive(b)
	return n, err
}

// SetDeadline implements the net.Conn SetDeadline method.
func (f *Win32File) SetDeadline(t time.Time) error {
	if err := f.SetReadDeadline(t); err != nil {
		return fmt.Errorf("set read deadline: %w", err)
	}
	if err := f.SetWriteDeadline(t); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}
	return nil
}

func (f *Win32File) SetReadDeadline(deadline time.Time) error {
	return f.ReadDeadline.set(deadline)
}

func (f *Win32File) SetWriteDeadline(deadline time.Time) error {
	return f.WriteDeadline.set(deadline)
}

func (f *Win32File) Flush() error {
	return syscall.FlushFileBuffers(f.Handle)
}

func (f *Win32File) IsSocket() bool {
	return f.isSocket
}

func (f *Win32File) Fd() uintptr {
	return uintptr(f.Handle)
}

// OSFile returns an [os.File] with the same underlying handle.
func (f *Win32File) OSFile(name string) *os.File {
	return os.NewFile(f.Fd(), name)
}

type emptyCh chan struct{}
type blockChan = <-chan struct{}

type deadline interface {
	Done() blockChan
}

// TODO:
// https://cs.opensource.google/go/go/+/refs/tags/go1.18.4:src/internal/poll/fd_poll_runtime.go;l=28;drc=bf2ef26be3593d24487311576d85ec601185fbf4;bpv=0;bpt=1
// https://cs.opensource.google/go/go/+/master:src/runtime/netpoll.go;drc=bf2ef26be3593d24487311576d85ec601185fbf4;l=103
// https://cs.opensource.google/go/go/+/master:src/runtime/netpoll.go;drc=bf2ef26be3593d24487311576d85ec601185fbf4;l=326

type deadlineHandler struct {
	setLock     sync.Mutex
	channel     emptyCh
	channelLock sync.RWMutex
	timer       *time.Timer
	timedout    isync.AtomicBool
}

var _ deadline = &deadlineHandler{}

func newDeadlineHandler() deadlineHandler {
	return deadlineHandler{
		channel: make(emptyCh),
	}
}

func (d *deadlineHandler) Done() blockChan {
	if d == nil {
		return make(emptyCh)
	}

	d.channelLock.Lock()
	defer d.channelLock.Unlock()
	return d.channel
}

func (d *deadlineHandler) set(deadline time.Time) error {
	d.setLock.Lock()
	defer d.setLock.Unlock()

	if d.timer != nil {
		if !d.timer.Stop() {
			<-d.channel
		}
		d.timer = nil
	}
	d.timedout.SetFalse()

	select {
	case <-d.channel:
		d.channelLock.Lock()
		d.channel = make(emptyCh)
		d.channelLock.Unlock()
	default:
	}

	if deadline.IsZero() {
		return nil
	}

	timeoutIO := func() {
		d.timedout.SetTrue()
		close(d.channel)
	}

	now := time.Now()
	duration := deadline.Sub(now)
	if deadline.After(now) {
		// Deadline is in the future, set a timer to wait
		d.timer = time.AfterFunc(duration, timeoutIO)
	} else {
		// Deadline is in the past. Cancel all pending IO now.
		timeoutIO()
	}
	return nil
}
