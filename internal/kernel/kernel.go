//go:build windows

package kernel

//go:generate go run github.com/Microsoft/go-winio/tools/mkwinsyscall -import github.com/Microsoft/go-winio/internal/wintype -output zkernel.go kernel.go

// DECLSPEC_ALLOCATOR HLOCAL LocalAlloc(
//   [in] UINT   uFlags,
//   [in] SIZE_T uBytes
// );
//
//sys LocalAlloc(flags wintype.UInt, bytes wintype.SizeT) (mem wintype.HLocal, err error)
//sys LocalAllocP(flags wintype.UInt, bytes wintype.SizeT) (mem wintype.HLocal, panic error) = LocalAlloc

// DECLSPEC_ALLOCATOR HLOCAL LocalReAlloc(
//   [in] _Frees_ptr_opt_ HLOCAL hMem,
//   [in] SIZE_T                 uBytes,
//   [in] UINT                   uFlags
// );
//
//sys LocalReAlloc(hmem wintype.HLocal, bytes wintype.SizeT, flags wintype.UInt) (mem wintype.HLocal, err error)
//sys LocalReAllocP(hmem wintype.HLocal, bytes wintype.SizeT, flags wintype.UInt) (mem wintype.HLocal, panic error) = LocalReAlloc

// HLOCAL LocalFree(
//   [in] _Frees_ptr_opt_ HLOCAL hMem
// );
//
//sys LocalFree(hmem wintype.HLocal) (mem wintype.HLocal, err error) [failretval!=0]
//sys LocalFreeP(hmem wintype.HLocal) (mem wintype.HLocal, panic error) [failretval!=0] = LocalFree
