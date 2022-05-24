// This package defines type aliases (ie, `#define`s) for windows data types.
// The intent is for translating API definitions, not to enforce type safety.
//
// See:
//  - https://docs.microsoft.com/en-us/windows/win32/winprog/windows-data-types
//  - https://docs.microsoft.com/en-us/cpp/cpp/data-type-ranges
package wintype

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
	HLocal    = uintptr // HANDLE
	HResult   = int32   // LONG
	Int       = int32   // int
	IntPtr    = int     // int or __int64: 32 or 64 bit
	Long32    = int32   // LONG32
	Long      = int32   // long
	LongLong  = int64   // long long
	LongPtr   = int     // long or __int64L 32 or 64 bits
	LPCStr    = *uint8  // *CHAR
	LPWStr    = *uint16 // *WCHAR
	LPVoid    = uintptr // *void
	PVoid     = uintptr // *void
	PWStr     = uint16  // *WCHAR
	QWord     = uint64  // unsigned __int64
	Short     = int16   // short
	SizeT     = uint    // ULONG_PTR
	SSizeT    = int     // LONG_PTR
	UInt      = uint32  // unsigned int
	UIntPtr   = uint    // unsigned int or unsigned __int64: 32 or 64 bits
	ULongPtr  = uint    // unsigned int or unsigned __int64: 32 or 64 bits
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
)

// todo: move to `internal/socket` when https://github.com/microsoft/go-winio/pull/239 is merged
// https://docs.microsoft.com/en-us/windows/win32/winsock/socket-data-type-2
// https://docs.microsoft.com/en-us/windows/win32/winsock/handling-winsock-errors

type Socket = UInt

const (
	SocketError   = ^Socket(0)
	InvalidSocket = Socket(0xffff)
)
