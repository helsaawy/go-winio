package wintype

// Microsoft C/C++ compiler types

//nolint:revive,stylecheck
type (
	C__ptr32    = uint32  // __ptr32
	C__ptr64    = uint64  // __ptr64
	C__wchar_t  = uint16  // __wchar_t
	Cchar       = int8    // char
	Cdouble     = float64 // double
	Cfloat      = float32 // float
	Cint        = int32   // int
	Clongdouble = float64 // long double
	Clong       = int32   // long
	Clongint    = int32   // long int
	Clonglong   = int64   // long long
	Cshort      = int16   // short
	Cshortint   = int16   // short int
	Cwchar_t    = uint16  // wchar_t
)
