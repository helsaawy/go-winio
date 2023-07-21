//go:build windows

package fs

import (
	"errors"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sys/windows"
)

//sys cancelIoEx(file windows.Handle, o *windows.Overlapped) (err error) = CancelIoEx
//sys createIoCompletionPort(file windows.Handle, port windows.Handle, key uintptr, threadCount uint32) (newport windows.Handle, err error) = CreateIoCompletionPort
//sys getQueuedCompletionStatus(port windows.Handle, bytes *uint32, key *uintptr, o **ioOperation, timeout uint32) (err error) = GetQueuedCompletionStatus
//sys setFileCompletionNotificationModes(h windows.Handle, flags uint8) (err error) = SetFileCompletionNotificationModes
//sys wsaGetOverlappedResult(h windows.Handle, o *windows.Overlapped, bytes *uint32, wait bool, flags *uint32) (err error) = ws2_32.WSAGetOverlappedResult

//todo (go1.19): switch to [atomic.Bool]

type AtomicBool int32

func (b *AtomicBool) isSet() bool { return atomic.LoadInt32((*int32)(b)) != 0 }
func (b *AtomicBool) setFalse()   { atomic.StoreInt32((*int32)(b), 0) }
func (b *AtomicBool) setTrue()    { atomic.StoreInt32((*int32)(b), 1) }

//revive:disable-next-line:predeclared Keep "new" to maintain consistency with "atomic" pkg
func (b *AtomicBool) swap(new bool) bool {
	var newInt int32
	if new {
		newInt = 1
	}
	return atomic.SwapInt32((*int32)(b), newInt) == 1
}

var (
	ErrFileClosed = errors.New("file has already been closed")
	ErrTimeout    = &timeoutError{}
)

type timeoutError struct{}

func (*timeoutError) Error() string   { return "i/o timeout" }
func (*timeoutError) Timeout() bool   { return true }
func (*timeoutError) Temporary() bool { return true }

type timeoutChan chan struct{}

var ioInitOnce sync.Once
var ioCompletionPort windows.Handle

// IOResult contains the result of an asynchronous IO operation.
type IOResult struct {
	Bytes uint32
	Err   error
}

// IOOperation represents an outstanding asynchronous Win32 IO.
type IOOperation struct {
	O windows.Overlapped
	C chan IOResult
}

func initIO() {
	h, err := createIoCompletionPort(windows.InvalidHandle, 0, 0, 0xffffffff)
	if err != nil {
		panic(err)
	}
	ioCompletionPort = h
	go ioCompletionProcessor(h)
}

// File implements Reader, Writer, and Closer on a Win32 handle without blocking in a syscall.
// It takes ownership of this handle and will close it if it is garbage collected.
type File struct {
	Handle        windows.Handle
	WG            sync.WaitGroup
	WGLock        sync.RWMutex
	Closing       AtomicBool
	Socket        bool
	ReadDeadline  deadlineHandler
	WriteDeadline deadlineHandler
}

var _ io.ReadWriteCloser = (*File)(nil)

type deadlineHandler struct {
	setLock     sync.Mutex
	channel     timeoutChan
	channelLock sync.RWMutex
	timer       *time.Timer
	timedout    AtomicBool
}

// MakeFile makes a new File from an existing file handle.
func MakeFile(h windows.Handle) (*File, error) {
	f := &File{Handle: h}
	ioInitOnce.Do(initIO)
	_, err := createIoCompletionPort(h, ioCompletionPort, 0, 0xffffffff)
	if err != nil {
		return nil, err
	}
	err = setFileCompletionNotificationModes(h, windows.FILE_SKIP_COMPLETION_PORT_ON_SUCCESS|windows.FILE_SKIP_SET_EVENT_ON_HANDLE)
	if err != nil {
		return nil, err
	}
	f.ReadDeadline.channel = make(timeoutChan)
	f.WriteDeadline.channel = make(timeoutChan)
	return f, nil
}

// closeHandle closes the resources associated with a Win32 handle.
func (f *File) closeHandle() {
	f.WGLock.Lock()
	// Atomically set that we are closing, releasing the resources only once.
	if !f.Closing.swap(true) {
		f.WGLock.Unlock()
		// cancel all IO and wait for it to complete
		_ = cancelIoEx(f.Handle, nil)
		f.WG.Wait()
		// at this point, no new IO can start
		windows.Close(f.Handle)
		f.Handle = 0
	} else {
		f.WGLock.Unlock()
	}
}

// Close closes a File.
func (f *File) Close() error {
	f.closeHandle()
	return nil
}

// IsClosed checks if the file has been closed.
func (f *File) IsClosed() bool {
	return f.Closing.isSet()
}

// PrepareIO prepares for a new IO operation.
// The caller must call f.wg.Done() when the IO is finished, prior to Close() returning.
func (f *File) PrepareIO() (*IOOperation, error) {
	f.WGLock.RLock()
	if f.Closing.isSet() {
		f.WGLock.RUnlock()
		return nil, ErrFileClosed
	}
	f.WG.Add(1)
	f.WGLock.RUnlock()
	c := &IOOperation{}
	c.C = make(chan IOResult)
	return c, nil
}

// ioCompletionProcessor processes completed async IOs forever.
func ioCompletionProcessor(h windows.Handle) {
	for {
		var bytes uint32
		var key uintptr
		var op *IOOperation
		err := getQueuedCompletionStatus(h, &bytes, &key, &op, windows.INFINITE)
		if op == nil {
			panic(err)
		}
		op.C <- IOResult{bytes, err}
	}
}

// todo: helsaawy - create an asyncIO version that takes a context

// AsyncIO processes the return value from ReadFile or WriteFile, blocking until
// the operation has actually completed.
func (f *File) AsyncIO(c *IOOperation, d *deadlineHandler, bytes uint32, err error) (int, error) {
	if err != windows.ERROR_IO_PENDING { //nolint:errorlint // err is Errno
		return int(bytes), err
	}

	if f.Closing.isSet() {
		_ = cancelIoEx(f.Handle, &c.O)
	}

	var timeout timeoutChan
	if d != nil {
		d.channelLock.Lock()
		timeout = d.channel
		d.channelLock.Unlock()
	}

	var r IOResult
	select {
	case r = <-c.C:
		err = r.Err
		if err == windows.ERROR_OPERATION_ABORTED { //nolint:errorlint // err is Errno
			if f.Closing.isSet() {
				err = ErrFileClosed
			}
		} else if err != nil && f.Socket {
			// err is from Win32. Query the overlapped structure to get the winsock error.
			var bytes, flags uint32
			err = wsaGetOverlappedResult(f.Handle, &c.O, &bytes, false, &flags)
		}
	case <-timeout:
		_ = cancelIoEx(f.Handle, &c.O)
		r = <-c.C
		err = r.Err
		if err == windows.ERROR_OPERATION_ABORTED { //nolint:errorlint // err is Errno
			err = ErrTimeout
		}
	}

	// runtime.KeepAlive is needed, as c is passed via native
	// code to ioCompletionProcessor, c must remain alive
	// until the channel read is complete.
	// todo: (de)allocate *ioOperation via win32 heap functions, instead of needing to KeepAlive?
	runtime.KeepAlive(c)
	return int(r.Bytes), err
}

// Read reads from a file handle.
func (f *File) Read(b []byte) (int, error) {
	c, err := f.PrepareIO()
	if err != nil {
		return 0, err
	}
	defer f.WG.Done()

	if f.ReadDeadline.timedout.isSet() {
		return 0, ErrTimeout
	}

	var bytes uint32
	err = windows.ReadFile(f.Handle, b, &bytes, &c.O)
	n, err := f.AsyncIO(c, &f.ReadDeadline, bytes, err)
	runtime.KeepAlive(b)

	// Handle EOF conditions.
	if err == nil && n == 0 && len(b) != 0 {
		return 0, io.EOF
	} else if err == windows.ERROR_BROKEN_PIPE { //nolint:errorlint // err is Errno
		return 0, io.EOF
	} else {
		return n, err
	}
}

// Write writes to a file handle.
func (f *File) Write(b []byte) (int, error) {
	c, err := f.PrepareIO()
	if err != nil {
		return 0, err
	}
	defer f.WG.Done()

	if f.WriteDeadline.timedout.isSet() {
		return 0, ErrTimeout
	}

	var bytes uint32
	err = windows.WriteFile(f.Handle, b, &bytes, &c.O)
	n, err := f.AsyncIO(c, &f.WriteDeadline, bytes, err)
	runtime.KeepAlive(b)
	return n, err
}

func (f *File) SetReadDeadline(deadline time.Time) error {
	return f.ReadDeadline.set(deadline)
}

func (f *File) SetWriteDeadline(deadline time.Time) error {
	return f.WriteDeadline.set(deadline)
}

func (f *File) Flush() error {
	return windows.FlushFileBuffers(f.Handle)
}

func (f *File) Fd() uintptr {
	return uintptr(f.Handle)
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
	d.timedout.setFalse()

	select {
	case <-d.channel:
		d.channelLock.Lock()
		d.channel = make(chan struct{})
		d.channelLock.Unlock()
	default:
	}

	if deadline.IsZero() {
		return nil
	}

	timeoutIO := func() {
		d.timedout.setTrue()
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
