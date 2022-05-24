// This package defines type aliases (ie, `#define`s) for Windows-specific and
// underlying Microsoft C & C++ compiler data types.
// The intent is for translating API definitions, and not to enforce type safety.
// Whenever possible, use [golang.org/x/sys/windows] types (eg, [golang.org/x/sys/windows.Handle])
// to maintain compatibility with other packages and functions.
//
// See:
//   - [Windows Data Types]
//   - [Data Type Ranges]
//   - [__ptr32, __ptr64]
//
// [Windows Data Types]: https://docs.microsoft.com/en-us/windows/win32/winprog/windows-data-types
// [Data Type Ranges]: https://docs.microsoft.com/en-us/cpp/cpp/data-type-ranges
// [__ptr32, __ptr64]: https://docs.microsoft.com/en-us/cpp/cpp/ptr32-ptr64
package wintype
