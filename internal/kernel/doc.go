// This package exposes a more low-level API for functions found in KernelBase and Kernel32,
// without relying on "golang.org/x/sys/windows".
package kernel

// todo:
// add generic versions of LocalAlloc/Free (eg, func Alloc [T any] (...) (mem *T, error)
// heap functions: https://docs.microsoft.com/en-us/windows/win32/api/heapapi/nf-heapapi-heapcreate

//go:generate go run github.com/Microsoft/go-winio/tools/mkwinsyscall -import github.com/Microsoft/go-winio/internal/wintype -output zkernel.go kernel.go file.go
