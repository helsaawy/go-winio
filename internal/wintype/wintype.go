package wintype

// Windows specific types

type (
	Atom      = uint16  // WORD
	Bool      = int32   // int
	Boolean   = uint8   // BYTE
	CChar     = int8    // char: can be uint8 if compiled with /J
	Char      = int8    // char: can be uint8 if compiled with /J
	DWord     = uint32  // unsigned long
	DWordLong = uint64  // unsigned __int64
	Handle    = uintptr // PVOID
	HFile     = int32   // int
	HKey      = uintptr // HANDLE
	HGlobal   = uintptr // HANDLE
	HLocal    = uintptr // HANDLE
	HResult   = int32   // LONG
	Int       = int32   // int
	IntPtr    = int     // Signed integer type for pointer precision: __int64 on 64-bit; int on 32-bit
	Long32    = int32   // LONG32
	Long      = int32   // long
	LongLong  = int64   // long long
	LongPtr   = int     // Signed long type for pointer precision: __int64L on 64-bit; long on 32- or 64-bit
	LPCStr    = *uint8  // CHAR *
	LPCVoid   = uintptr // CONST void *: cannot declar type as const
	LPWStr    = *uint16 // WCHAR *
	LPVoid    = uintptr // void *
	Pointer32 = uint32  // __ptr32: truncated 64-bit pointer on 64-bit; native pointer on 32-bit
	Pointer64 = uint64  // __ptr64: native pointer on 64-bit; sign-extended 32-bit pointer on 32-bit
	PVoid     = uintptr // void *
	PWStr     = uint16  // WCHAR *
	QWord     = uint64  // unsigned __int64
	Short     = int16   // short
	SizeT     = uint    // ULONG_PTR
	SSizeT    = int     // LONG_PTR
	UInt      = uint32  // unsigned int
	UIntPtr   = uint    // Unsigned integer type for pointer precision: unsigned __int64 on 64-bit; unsigned int on 32-bit
	ULongLong = uint64  // unsigned long long
	ULongPtr  = uint    // Unsigned long type for pointer precision: unsigned __int64L on 64-bit; unsigned long on 32-bit
	UShort    = uint16  // unsigned short
	WChar     = uint16  // wchar_t
	Word      = uint16  // unsigned short
)

const (
	NullHandle    = Handle(0)
	NullPtr       = PVoid(0)
	InvalidHandle = ^Handle(0)

	True  = 1
	False = 0

	NULL = 0
)

// todo: move to `internal/socket` when https://github.com/microsoft/go-winio/pull/239 is merged
// https://docs.microsoft.com/en-us/windows/win32/winsock/socket-data-type-2
// https://docs.microsoft.com/en-us/windows/win32/winsock/handling-winsock-errors

type Socket = UInt

const (
	SocketError   = ^Socket(0)
	InvalidSocket = Socket(0xffff)
)
