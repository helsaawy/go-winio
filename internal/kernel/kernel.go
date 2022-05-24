//go:build windows

package kernel

import "github.com/Microsoft/go-winio/internal/wintype"

// BOOL CloseHandle(
//   [in] HANDLE hObject
// );
//
//sys CloseHandle(h wintype.Handle) (ok bool, err error) [!failretval]

func CloseHandleMust(h wintype.Handle) {
	if _, err := CloseHandle(h); err != nil {
		panic(err)
	}
}

//
// global stack allocations
//

// DECLSPEC_ALLOCATOR HGLOBAL GlobalAlloc(
//   [in] UINT   uFlags,
//   [in] SIZE_T dwBytes
// );
//
//sys GlobalAlloc(flags wintype.UInt, bytes wintype.SizeT) (mem wintype.HGlobal, err error)

func GlobalAllocMust(flags wintype.UInt, bytes wintype.SizeT) wintype.HGlobal {
	mem, err := GlobalAlloc(flags, bytes)
	if err != nil {
		panic(err)
	}
	return mem
}

// DECLSPEC_ALLOCATOR HGLOBAL GlobalReAlloc(
//   [in] _Frees_ptr_ HGLOBAL hMem,
//   [in] SIZE_T              dwBytes,
//   [in] UINT                uFlags
// );
//
//sys GlobalReAlloc(hMem wintype.HGlobal, bytes wintype.SizeT, flags wintype.UInt) (mem wintype.HGlobal, err error)

func GlobalReAllocMust(hMem wintype.HGlobal, bytes wintype.SizeT, flags wintype.UInt) wintype.HGlobal {
	mem, err := GlobalReAlloc(hMem, bytes, flags)
	if err != nil {
		panic(err)
	}
	return mem
}

// HGLOBAL GlobalFree(
//   [in] _Frees_ptr_opt_ HGLOBAL hMem
// );
//
//sys GlobalFree(hMem wintype.HGlobal) (mem wintype.HGlobal, err error) [failretval!=0]

func GlobalFreeMust(hMem wintype.HGlobal) wintype.HGlobal {
	mem, err := GlobalFree(hMem)
	if err != nil {
		panic(err)
	}
	return mem
}

// HGLOBAL GlobalHandle(
//   [in] LPCVOID pMem
// );
//
//sys GlobalHandle(pMem wintype.LPCVoid) (hMem wintype.HGlobal, err error)

// LPVOID GlobalLock(
//   [in] HGLOBAL hMem
// );
//
//sys GlobalLock(hMem wintype.HGlobal) (pMem wintype.LPVoid, err error)

// BOOL GlobalUnlock(
//   [in] HGLOBAL hMem
// );
//
//sys GlobalUnlock(hMem wintype.HGlobal) (ok bool, err error) [!failretval]

//
// local stack allocations
//

// DECLSPEC_ALLOCATOR HLOCAL LocalAlloc(
//   [in] UINT   uFlags,
//   [in] SIZE_T uBytes
// );
//
//sys LocalAlloc(flags wintype.UInt, bytes wintype.SizeT) (mem wintype.HLocal, err error)

func LocalAllocMust(flags wintype.UInt, bytes wintype.SizeT) wintype.HLocal {
	mem, err := LocalAlloc(flags, bytes)
	if err != nil {
		panic(err)
	}
	return mem
}

// DECLSPEC_ALLOCATOR HLOCAL LocalReAlloc(
//   [in] _Frees_ptr_opt_ HLOCAL hMem,
//   [in] SIZE_T                 uBytes,
//   [in] UINT                   uFlags
// );
//
//sys LocalReAlloc(hMem wintype.HLocal, bytes wintype.SizeT, flags wintype.UInt) (mem wintype.HLocal, err error)

func LocalReAllocMust(hMem wintype.HLocal, bytes wintype.SizeT, flags wintype.UInt) wintype.HLocal {
	mem, err := LocalReAlloc(hMem, bytes, flags)
	if err != nil {
		panic(err)
	}
	return mem
}

// HLOCAL LocalFree(
//   [in] _Frees_ptr_opt_ HLOCAL hMem
// );
//
//sys LocalFree(hMem wintype.HLocal) (mem wintype.HLocal, err error) [failretval!=0]

func LocalFreeMust(hMem wintype.HLocal) wintype.HLocal {
	mem, err := LocalFree(hMem)
	if err != nil {
		panic(err)
	}
	return mem
}

// HLOCAL LocalHandle(
//   [in] LPCVOID pMem
// );
//
//sys LocalHandle(pMem wintype.LPCVoid) (hMem wintype.HLocal, err error)

// LPVOID LocalLock(
//   [in] HLOCAL hMem
// );
//
//sys LocalLock(hMem wintype.HLocal) (pMem wintype.LPVoid, err error)

// BOOL LocalUnlock(
//   [in] HLOCAL hMem
// );
//
//sys LocalUnlock(hMem wintype.HLocal) (ok bool, err error) [!failretval]
