package file

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// see
// https://docs.microsoft.com/en-us/windows/win32/secauthz/access-mask
// https://docs.microsoft.com/en-us/windows/win32/fileio/file-access-rights-constants

type AccessMask uint32

const (
	SpecificRightsAll AccessMask = 0x0000FFFF

	FileReadData           AccessMask = 0x0001
	FileListDirectory      AccessMask = 0x0001
	FileWriteData          AccessMask = 0x0002
	FileAddFile            AccessMask = 0x0002
	FileAppendData         AccessMask = 0x0004
	FileAddSubdirectory    AccessMask = 0x0004
	FileCreatePipeInstance AccessMask = 0x0004
	FileReadEA             AccessMask = 0x0008
	FileWriteEA            AccessMask = 0x0010
	FileExecute            AccessMask = 0x0020
	FileTraverse           AccessMask = 0x0020
	FileDeleteChild        AccessMask = 0x0040
	FileReadAttributes     AccessMask = 0x0080
	FileWriteAttributes    AccessMask = 0x0100
	FileAllAccess          AccessMask = (StandardRightsRequired | Synchronize | 0x1FF)

	StandardRightsMask     AccessMask = 0x001F0000
	StandardRightsRequired AccessMask = 0x000F0000
	Delete                 AccessMask = 0x00010000
	ReadControl            AccessMask = 0x00020000
	WriteDAC               AccessMask = 0x00040000
	WriteOwner             AccessMask = 0x00080000
	Synchronize            AccessMask = 0x00100000
	StandardRightsRead     AccessMask = ReadControl
	StandardRightsWrite    AccessMask = ReadControl
	StandardRightsExecute  AccessMask = ReadControl

	AccessSystemSecurity AccessMask = 0x01000000
	MaximumAllowed       AccessMask = 0x02000000

	GenericAll     AccessMask = 0x10000000
	GenericExecute AccessMask = 0x20000000
	GenericWrite   AccessMask = 0x40000000
	GenericRead    AccessMask = 0x80000000
)

type ShareMode uint32

const (
	ShareNone       ShareMode = 0x0
	ShareRead       ShareMode = 0x1
	ShareWrite      ShareMode = 0x2
	ShareDelete     ShareMode = 0x4
	ShareValidFlags ShareMode = 0x7
)

type CreationDisposition uint32

const (
	CreateNew        CreationDisposition = 1
	CreateAlways     CreationDisposition = 2
	OpenExisting     CreationDisposition = 3
	OpenAlways       CreationDisposition = 4
	TruncateExisting CreationDisposition = 5
)

// https://docs.microsoft.com/en-us/windows/win32/fileio/file-attribute-constants

type FlagOrAttribute uint32

const (
	AttributeReadonly           FlagOrAttribute = 0x00000001
	AttributeHidden             FlagOrAttribute = 0x00000002
	AttributeSystem             FlagOrAttribute = 0x00000004
	AttributeDirectory          FlagOrAttribute = 0x00000010
	AttributeArchive            FlagOrAttribute = 0x00000020
	AttributeDevice             FlagOrAttribute = 0x00000040
	AttributeNormal             FlagOrAttribute = 0x00000080
	AttributeTemporary          FlagOrAttribute = 0x00000100
	AttributeSparseFile         FlagOrAttribute = 0x00000200
	AttributeReparsePoint       FlagOrAttribute = 0x00000400
	AttributeCompressed         FlagOrAttribute = 0x00000800
	AttributeOffline            FlagOrAttribute = 0x00001000
	AttributeNotContentIndexed  FlagOrAttribute = 0x00002000
	AttributeEncrypted          FlagOrAttribute = 0x00004000
	AttributeIntegrityStream    FlagOrAttribute = 0x00008000
	AttributeVirtual            FlagOrAttribute = 0x00010000
	AttributeNoScrubData        FlagOrAttribute = 0x00020000
	AttributeEA                 FlagOrAttribute = 0x00040000
	AttributeRecallOnOpen       FlagOrAttribute = 0x00040000
	AttributeRecallOnDataAccess FlagOrAttribute = 0x00400000
	AttributePinned             FlagOrAttribute = 0x00080000
	AttributeUnpinned           FlagOrAttribute = 0x00100000
	AttributeStrictlySequential FlagOrAttribute = 0x20000000

	FlagWriteThrough        FlagOrAttribute = 0x80000000
	FlagOverlapped          FlagOrAttribute = 0x40000000
	FlagNoBuffering         FlagOrAttribute = 0x20000000
	FlagRandomAccess        FlagOrAttribute = 0x10000000
	FlagSequentialScan      FlagOrAttribute = 0x08000000
	FlagDeleteOnClose       FlagOrAttribute = 0x04000000
	FlagBackupSemantics     FlagOrAttribute = 0x02000000
	FlagPosixSemantics      FlagOrAttribute = 0x01000000
	FlagSessionAware        FlagOrAttribute = 0x00800000
	FlagOpenReparsePoint    FlagOrAttribute = 0x00200000
	FlagOpenNoRecall        FlagOrAttribute = 0x00100000
	FlagFirstPipeInstance   FlagOrAttribute = 0x00080000
	FlagOpenRequiringOplock FlagOrAttribute = 0x00040000

	SecurityAnonymous       FlagOrAttribute = 0
	SecurityIdentification  FlagOrAttribute = 0x00010000
	SecurityImpersonation   FlagOrAttribute = 0x00020000
	SecurityDelegation      FlagOrAttribute = 0x00030000
	SecurityContextTracking FlagOrAttribute = 0x00040000
	SecurityEffectiveOnly   FlagOrAttribute = 0x00080000
	SecuritySQOSPresent     FlagOrAttribute = 0x00100000
	SecurityValidSQOSFlags  FlagOrAttribute = 0x001F0000
)

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
//sys CreateFile(name string, access AccessMask, share ShareMode, sa *windows.SecurityAttributes, mode CreationDisposition, attrs FlagOrAttribute, template windows.Handle) (handle windows.Handle, err error) [failretval==windows.InvalidHandle] = CreateFileW

// OpenFile opens (or creates, if it does not exits) a file for overlapped (async) IO.
// It is a simplified version of [CreateFile], and will set [FlagOverlapped] if it is not
// already enabled.
func OpenFile(file string, access AccessMask, attr FlagOrAttribute) (*Win32File, error) {
	attr |= FlagOverlapped
	h, err := CreateFile(file, access, ShareRead|ShareWrite, inheritableSA(), CreateAlways, attr, 0)
	if err != nil {
		return nil, err
	}
	return MakeWin32File(syscall.Handle(h), false)
}

func inheritableSA() *windows.SecurityAttributes {
	const s = unsafe.Sizeof(windows.SecurityAttributes{})
	return &windows.SecurityAttributes{
		InheritHandle: 1,
		Length:        uint32(s),
	}
}
