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
	CreateNew        = 1
	CreateAlways     = 2
	OpenExisting     = 3
	OpenAlways       = 4
	TruncateExisting = 5
)

// https://docs.microsoft.com/en-us/windows/win32/fileio/file-attribute-constants

type FlagOrAttribute uint32

const (
	AttributeReadonly           = 0x00000001
	AttributeHidden             = 0x00000002
	AttributeSystem             = 0x00000004
	AttributeDirectory          = 0x00000010
	AttributeArchive            = 0x00000020
	AttributeDevice             = 0x00000040
	AttributeNormal             = 0x00000080
	AttributeTemporary          = 0x00000100
	AttributeSparseFile         = 0x00000200
	AttributeReparsePoint       = 0x00000400
	AttributeCompressed         = 0x00000800
	AttributeOffline            = 0x00001000
	AttributeNotContentIndexed  = 0x00002000
	AttributeEncrypted          = 0x00004000
	AttributeIntegrityStream    = 0x00008000
	AttributeVirtual            = 0x00010000
	AttributeNoScrubData        = 0x00020000
	AttributeEA                 = 0x00040000
	AttributeRecallOnOpen       = 0x00040000
	AttributeRecallOnDataAccess = 0x00400000
	AttributePinned             = 0x00080000
	AttributeUnpinned           = 0x00100000
	AttributeStrictlySequential = 0x20000000

	FlagWriteThrough        = 0x80000000
	FlagOverlapped          = 0x40000000
	FlagNoBuffering         = 0x20000000
	FlagRandomAccess        = 0x10000000
	FlagSequentialScan      = 0x08000000
	FlagDeleteOnClose       = 0x04000000
	FlagBackupSemantics     = 0x02000000
	FlagPosixSemantics      = 0x01000000
	FlagSessionAware        = 0x00800000
	FlagOpenReparsePoint    = 0x00200000
	FlagOpenNoRecall        = 0x00100000
	FlagFirstPipeInstance   = 0x00080000
	FlagOpenRequiringOplock = 0x00040000

	SecurityAnonymous       = 0
	SecurityIdentification  = 0x00010000
	SecurityImpersonation   = 0x00020000
	SecurityDelegation      = 0x00030000
	SecurityContextTracking = 0x00040000
	SecurityEffectiveOnly   = 0x00080000
	SecuritySQOSPresent     = 0x00100000
	SecurityValidSQOSFlags  = 0x001F0000
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
//sys CreateFile(name string, access AccessMask, mode ShareMode, sa *windows.SecurityAttributes, createmode CreationDisposition, attrs FlagOrAttribute, template windows.Handle) (handle windows.Handle, err error) [failretval==windows.InvalidHandle] = CreateFileW

// OpenFile opens (or creates, depending on the flags passed) the file specified by path.
// It is a simplified version of [CreateFile].
func OpenFile(path string, access AccessMask, attr FlagOrAttribute) (*Win32File, error) {
	h, err := CreateFile(path, access, ShareRead|ShareWrite, inheritableSA(), CreateNew, attr, 0)
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
