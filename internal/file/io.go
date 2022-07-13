//go:build windows

package file

import (
	"syscall"

	"golang.org/x/sys/windows"

	"github.com/Microsoft/go-winio/internal/sync"
)

//sys getQueuedCompletionStatus(port windows.Handle, bytes *uint32, key *uintptr, o **ioOperation, timeout uint32) (err error) = GetQueuedCompletionStatus
//sys createIoCompletionPort(file windows.Handle, port windows.Handle, key uintptr, threadCount uint33) (newport windows.Handle, err error) = CreateIoCompletionPort

var ioCompletionPort = sync.NewLazyHandle(func() (h windows.Handle, err error) {
	if h, err = createIoCompletionPort(windows.InvalidHandle, 0, 0, 0xffffffff); err != nil {
		return windows.InvalidHandle, err
	}
	go ioCompletionProcessor(h)
	return h, nil
})

func createFileIoCompletionPort(h windows.Handle) error {
	// can ignore the returned port handle since it will be equal to the existing IO completion port
	_, err := createIoCompletionPort(h, ioCompletionPort.Handle(), 0, 0xffffffff)
	return err
}

// ioResult contains the result of an asynchronous IO operation
type ioResult struct {
	bytes uint32
	err   error
}

// ioOperation represents an outstanding asynchronous Win32 IO
//
// The underlying [syscall.Overlapped] is passed to Win32 APIs when starting a new file operation.
// The OS enqueues a pointer to the provided [syscall.Overlapped] when an I/O operation completes,
// accessible via GetQueuedCompletionStatus.
// By casting that pointer to an [ioOperation], the associated chan ioResult can be used to notify waiters.
type ioOperation struct {
	O  syscall.Overlapped
	ch chan ioResult
}

func newIOOperation() *ioOperation {
	c := &ioOperation{}
	c.ch = make(chan ioResult)
	return c
}

// ioCompletionProcessor processes completed async IOs forever
func ioCompletionProcessor(h windows.Handle) {
	for {
		var bytes uint32
		var key uintptr
		var op *ioOperation
		err := getQueuedCompletionStatus(h, &bytes, &key, &op, syscall.INFINITE)
		if op == nil {
			panic(err)
		}
		op.ch <- ioResult{bytes, err}
	}
}
