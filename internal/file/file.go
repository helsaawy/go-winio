//go:build windows

package file

// HANDLE CreateFileW(
//   [in]           LPCWSTR               lpFileName,
//   [in]           DWORD                 dwDesiredAccess,
//   [in]           DWORD                 dwShareMode,
//   [in, optional] LPSECURITY_ATTRIBUTES lpSecurityAttributes,
//   [in]           DWORD                 dwCreationDisposition,
//   [in]           DWORD                 dwFlagsAndAttributes,
//   [in, optional] HANDLE                hTemplateFile
// );
//
//sys CreateFile(name string, access wintype.DWord, mode wintype.DWord, sa *windows.SecurityAttributes, createmode wintype.DWord, attrs wintype.DWord, templatefile windows.Handle) (handle windows.Handle, err error) [failretval==windows.InvalidHandle] = CreateFileW
