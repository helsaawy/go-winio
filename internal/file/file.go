//go:build windows

package file

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/windows"

	"github.com/Microsoft/go-winio/internal/deadline"
	isync "github.com/Microsoft/go-winio/internal/sync"
)

//sys cancelIoEx(file syscall.Handle, o *syscall.Overlapped) (err error) = CancelIoEx
//sys setFileCompletionNotificationModes(h syscall.Handle, flags uint8) (err error) = SetFileCompletionNotificationModes
//sys wsaGetOverlappedResult(h syscall.Handle, o *syscall.Overlapped, bytes *uint32, wait bool, flags *uint32) (err error) = ws2_32.WSAGetOverlappedResult

const (
	skipCompletionPortOnSuccess = 1
	skipSetEventOnHandle        = 2
)

var ErrFileClosed = os.ErrClosed

// todo: ReadFrom, ReadAt, WriteAt, WriteString, Seek
// todo: GetFinalPathNameByHandleA and get file name from handle; win32file -> path -> os.Open

// Win32File implements Reader, Writer, and Closer on a Win32 handle without blocking in a syscall.
// It takes ownership of this handle and will close it if it is garbage collected.
type Win32File struct {
	Handle        syscall.Handle
	ReadDeadline  *deadline.Deadline
	WriteDeadline *deadline.Deadline

	isSocket bool
	// wgLock should be locked when closing or adding new operations to prevent new operations
	// from starting while mutating file state
	wgLock sync.RWMutex
	// wg is the group for all pending IO operations
	wg      sync.WaitGroup
	closing isync.AtomicBool
}

var _ io.ReadWriteCloser = &Win32File{}

// MakeWin32File makes a new win32File from an existing file handle (eg, from the result of [CreateFile]).
func MakeWin32File(h syscall.Handle, socket bool) (*Win32File, error) {
	f := &Win32File{
		Handle:        h,
		ReadDeadline:  deadline.Empty(),
		WriteDeadline: deadline.Empty(),
	}
	if err := createFileIoCompletionPort(windows.Handle(h)); err != nil {
		return nil, fmt.Errorf("create file IO completion port: %w", err)
	}
	if err := setFileCompletionNotificationModes(h, skipCompletionPortOnSuccess|skipSetEventOnHandle); err != nil {
		return nil, fmt.Errorf("set file IO notification mode: %w", err)
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
func (f *Win32File) AsyncIo(d deadline.Timeout, c *IoOperation, bytes uint32, err error) (int, error) {
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
		if errnoIs(err, windows.ERROR_OPERATION_ABORTED) {
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
		if errnoIs(err, windows.ERROR_OPERATION_ABORTED) {
			err = deadline.ErrDeadlineExceeded
		}
	}

	// runtime.KeepAlive is needed, as c is passed via native
	// code to ioCompletionProcessor, c must remain alive
	// until the channel read is complete.
	// todo: allocate *ioOperation via win32 heap functions, instead of needing to KeepAlive?
	runtime.KeepAlive(c)
	return int(r.bytes), err
}

//todo: create `StartRead([]byte) <-chan IOResult` (and `StartWrite(`  version) that expose async operations

// Read reads from a file handle.
func (f *Win32File) Read(b []byte) (int, error) {
	c, err := f.PrepareIo()
	if err != nil {
		return 0, err
	}
	defer c.Close()

	if err = f.ReadDeadline.Err(); err != nil {
		return 0, err
	}

	var bytes uint32
	err = syscall.ReadFile(f.Handle, b, &bytes, &c.O)
	n, err := f.AsyncIo(f.ReadDeadline, c, bytes, err)
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

	if err = f.WriteDeadline.Err(); err != nil {
		return 0, err
	}

	var bytes uint32
	err = syscall.WriteFile(f.Handle, b, &bytes, &c.O)
	n, err := f.AsyncIo(f.WriteDeadline, c, bytes, err)
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
	return f.ReadDeadline.Reset(deadline)
}

func (f *Win32File) SetWriteDeadline(deadline time.Time) error {
	return f.WriteDeadline.Reset(deadline)
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

// errors.Is, but specialized for windows.Errorno
func errnoIs(err error, target windows.Errno) bool {
	n, ok := err.(windows.Errno) //nolint:errorlint
	return ok && n == target
}
