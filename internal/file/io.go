//go:build windows

package file

import (
	"fmt"
	"sync"
	"syscall"

	"golang.org/x/sys/windows"
)

//sys getQueuedCompletionStatus(port windows.Handle, bytes *uint32, key *uintptr, o **IoOperation, timeout uint32) (err error) = GetQueuedCompletionStatus
//sys createIoCompletionPort(file windows.Handle, port windows.Handle, key uintptr, threadCount uint32) (newport windows.Handle, err error) = CreateIoCompletionPort

// global processor
var _processor ioCompletionProcessor

type ioCompletionProcessor struct {
	// h is the I/O completion port
	h windows.Handle
	// once initializes h, and starts processing
	once sync.Once
}

func (p *ioCompletionProcessor) port() windows.Handle {
	p.init()
	return p.h
}

func (p *ioCompletionProcessor) init() {
	p.once.Do(func() {
		var err error
		p.h, err = createIoCompletionPort(windows.InvalidHandle, 0, 0, 0xffffffff)
		if err != nil {
			panic(fmt.Sprintf("could not create a new I/O completion port: %v", err))
		}
		go p.start()
	})
}

// start loops forever, notifying [IoOperations] when their I/O operation completes.
// Assumes that [ioCompletionQueueProcessor.do] has been called and [p.h] is initialized and valid.
func (p *ioCompletionProcessor) start() {
	for {
		var bytes uint32
		var key uintptr
		var op *IoOperation
		err := getQueuedCompletionStatus(p.h, &bytes, &key, &op, syscall.INFINITE)
		if op == nil {
			panic(err)
		}
		op.ch <- ioResult{bytes, err}
	}
}

func createFileIoCompletionPort(h windows.Handle) error {
	// can ignore the returned port handle since it will be equal to the existing IO completion port
	_, err := createIoCompletionPort(h, _processor.port(), 0, 0xffffffff)
	return err
}

// IoOperation represents an outstanding asynchronous (overlapped) Win32 I/O operation.
//
// The underlying [syscall.Overlapped] is passed to Win32 APIs when starting a new file operation.
// The OS enqueues a pointer to the provided [syscall.Overlapped] when an I/O operation completes,
// accessible via GetQueuedCompletionStatus.
// By casting that pointer to an [IoOperation], the associated chan ioResult can be used to notify waiters.
type IoOperation struct {
	O  syscall.Overlapped
	ch chan ioResult
	f  *Win32File
}

// newIoOperation creates an [IoOperation] associated with the [Win32File] f.
// The caller must hold f.wgLock.RLock
func newIoOperation(f *Win32File) *IoOperation {
	c := &IoOperation{
		f: f,
	}
	f.wg.Add(1)
	c.ch = make(chan ioResult)
	return c
}

func (c *IoOperation) Close() error {
	close(c.ch)
	c.f.wg.Done()
	c.f = nil
	return nil
}

// ioResult contains the result of an asynchronous IO operation
type ioResult struct {
	bytes uint32
	err   error
}
