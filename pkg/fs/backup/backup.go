//go:build windows

package backup

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"unicode/utf16"

	"golang.org/x/sys/windows"

	"github.com/Microsoft/go-winio/internal/fs"
)

//go:generate go run github.com/Microsoft/go-winio/tools/mkwinsyscall -output zsyscall_windows.go backup.go

//sys backupRead(h windows.Handle, b []byte, bytesRead *uint32, abort bool, processSecurity bool, context *uintptr) (err error) = BackupRead
//sys backupWrite(h windows.Handle, b []byte, bytesWritten *uint32, abort bool, processSecurity bool, context *uintptr) (err error) = BackupWrite

type StreamID uint32

const (
	BackupData = StreamID(iota + 1)
	BackupEAData
	BackupSecurity
	BackupAlternateData
	BackupLink
	BackupPropertyData
	BackupObjectID
	BackupReparseData
	BackupSparseBlock
	BackupTXFsData
)

type StreamAttribute uint32

const (
	StreamNormalAttribute    StreamAttribute = 0x0
	StreamModifiedWhenRead   StreamAttribute = 0x1
	StreamContainsSecurity   StreamAttribute = 0x2
	StreamContainsProperties StreamAttribute = 0x4
	StreamSparseAttribute    StreamAttribute = 0x8
)

// Header represents a backup stream of a file.
type Header struct {
	ID         StreamID        // The backup stream ID
	Attributes StreamAttribute // Stream attributes
	Size       int64           // The size of the stream in bytes
	Name       string          // The name of the stream (for BackupAlternateData only).
	Offset     int64           // The offset of the stream in the file (for BackupSparseBlock only).
}

//	typedef struct _WIN32_STREAM_ID {
//	  DWORD         dwStreamId;
//	  DWORD         dwStreamAttributes;
//	  LARGE_INTEGER Size;
//	  DWORD         dwStreamNameSize;
//	  WCHAR         cStreamName[ANYSIZE_ARRAY];
//	} WIN32_STREAM_ID, *LPWIN32_STREAM_ID;
//
// https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-win32_stream_id
type win32StreamID struct {
	StreamID   StreamID
	Attributes StreamAttribute
	Size       uint64
	NameSize   uint32
}

// StreamReader reads from a stream produced by the BackupRead Win32 API and produces a series
// of BackupHeader values.
type StreamReader struct {
	r         io.Reader
	bytesLeft int64
}

// NewStreamReader produces a BackupStreamReader from any io.Reader.
func NewStreamReader(r io.Reader) *StreamReader {
	return &StreamReader{r, 0}
}

// Next returns the next backup stream and prepares for calls to Read(). It skips the remainder of the current stream if
// it was not completely read.
func (r *StreamReader) Next() (*Header, error) {
	if r.bytesLeft > 0 { //nolint:nestif // todo: flatten this
		if s, ok := r.r.(io.Seeker); ok {
			// Make sure Seek on io.SeekCurrent sometimes succeeds
			// before trying the actual seek.
			if _, err := s.Seek(0, io.SeekCurrent); err == nil {
				if _, err = s.Seek(r.bytesLeft, io.SeekCurrent); err != nil {
					return nil, err
				}
				r.bytesLeft = 0
			}
		}
		if _, err := io.Copy(io.Discard, r); err != nil {
			return nil, err
		}
	}
	var wsi win32StreamID
	if err := binary.Read(r.r, binary.LittleEndian, &wsi); err != nil {
		return nil, err
	}
	hdr := &Header{
		ID:         wsi.StreamID,
		Attributes: wsi.Attributes,
		Size:       int64(wsi.Size),
	}
	if wsi.NameSize != 0 {
		name := make([]uint16, int(wsi.NameSize/2))
		if err := binary.Read(r.r, binary.LittleEndian, name); err != nil {
			return nil, err
		}
		hdr.Name = windows.UTF16ToString(name)
	}
	if wsi.StreamID == BackupSparseBlock {
		if err := binary.Read(r.r, binary.LittleEndian, &hdr.Offset); err != nil {
			return nil, err
		}
		hdr.Size -= 8
	}
	r.bytesLeft = hdr.Size
	return hdr, nil
}

// Read reads from the current backup stream.
func (r *StreamReader) Read(b []byte) (int, error) {
	if r.bytesLeft == 0 {
		return 0, io.EOF
	}
	if int64(len(b)) > r.bytesLeft {
		b = b[:r.bytesLeft]
	}
	n, err := r.r.Read(b)
	r.bytesLeft -= int64(n)
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	} else if r.bytesLeft == 0 && err == nil {
		err = io.EOF
	}
	return n, err
}

// StreamWriter writes a stream compatible with the BackupWrite Win32 API.
type StreamWriter struct {
	w         io.Writer
	bytesLeft int64
}

// NewStreamWriter produces a BackupStreamWriter on top of an io.Writer.
func NewStreamWriter(w io.Writer) *StreamWriter {
	return &StreamWriter{w, 0}
}

// WriteHeader writes the next backup stream header and prepares for calls to Write().
func (w *StreamWriter) WriteHeader(hdr *Header) error {
	if w.bytesLeft != 0 {
		return fmt.Errorf("missing %d bytes", w.bytesLeft)
	}
	name := utf16.Encode([]rune(hdr.Name))
	wsi := win32StreamID{
		StreamID:   hdr.ID,
		Attributes: hdr.Attributes,
		Size:       uint64(hdr.Size),
		NameSize:   uint32(len(name) * 2),
	}
	if wsi.StreamID == BackupSparseBlock {
		// Include space for the int64 block offset
		wsi.Size += 8
	}
	if err := binary.Write(w.w, binary.LittleEndian, &wsi); err != nil {
		return err
	}
	if len(name) != 0 {
		if err := binary.Write(w.w, binary.LittleEndian, name); err != nil {
			return err
		}
	}
	if wsi.StreamID == BackupSparseBlock {
		if err := binary.Write(w.w, binary.LittleEndian, hdr.Offset); err != nil {
			return err
		}
	}
	w.bytesLeft = hdr.Size
	return nil
}

// Write writes to the current backup stream.
func (w *StreamWriter) Write(b []byte) (int, error) {
	if w.bytesLeft < int64(len(b)) {
		return 0, fmt.Errorf("too many bytes by %d", int64(len(b))-w.bytesLeft)
	}
	n, err := w.w.Write(b)
	w.bytesLeft -= int64(n)
	return n, err
}

// FileReader provides an io.ReadCloser interface on top of the BackupRead Win32 API.
type FileReader struct {
	f               *os.File
	includeSecurity bool
	ctx             uintptr
}

// NewFileReader returns a new BackupFileReader from a file handle. If includeSecurity is true,
// Read will attempt to read the security descriptor of the file.
func NewFileReader(f *os.File, includeSecurity bool) *FileReader {
	r := &FileReader{f, includeSecurity, 0}
	return r
}

// Read reads a backup stream from the file by calling the Win32 API BackupRead().
func (r *FileReader) Read(b []byte) (int, error) {
	var bytesRead uint32
	err := backupRead(windows.Handle(r.f.Fd()), b, &bytesRead, false, r.includeSecurity, &r.ctx)
	if err != nil {
		return 0, &os.PathError{Op: "BackupRead", Path: r.f.Name(), Err: err}
	}
	runtime.KeepAlive(r.f)
	if bytesRead == 0 {
		return 0, io.EOF
	}
	return int(bytesRead), nil
}

// Close frees Win32 resources associated with the BackupFileReader. It does not close
// the underlying file.
func (r *FileReader) Close() error {
	if r.ctx != 0 {
		_ = backupRead(windows.Handle(r.f.Fd()), nil, nil, true, false, &r.ctx)
		runtime.KeepAlive(r.f)
		r.ctx = 0
	}
	return nil
}

// FileWriter provides an io.WriteCloser interface on top of the BackupWrite Win32 API.
type FileWriter struct {
	f               *os.File
	includeSecurity bool
	ctx             uintptr
}

// NewFileWriter returns a new BackupFileWriter from a file handle. If includeSecurity is true,
// Write() will attempt to restore the security descriptor from the stream.
func NewFileWriter(f *os.File, includeSecurity bool) *FileWriter {
	w := &FileWriter{f, includeSecurity, 0}
	return w
}

// Write restores a portion of the file using the provided backup stream.
func (w *FileWriter) Write(b []byte) (int, error) {
	var bytesWritten uint32
	err := backupWrite(windows.Handle(w.f.Fd()), b, &bytesWritten, false, w.includeSecurity, &w.ctx)
	if err != nil {
		return 0, &os.PathError{Op: "BackupWrite", Path: w.f.Name(), Err: err}
	}
	runtime.KeepAlive(w.f)
	if int(bytesWritten) != len(b) {
		return int(bytesWritten), errors.New("not all bytes could be written")
	}
	return len(b), nil
}

// Close frees Win32 resources associated with the BackupFileWriter. It does not
// close the underlying file.
func (w *FileWriter) Close() error {
	if w.ctx != 0 {
		_ = backupWrite(windows.Handle(w.f.Fd()), nil, nil, true, false, &w.ctx)
		runtime.KeepAlive(w.f)
		w.ctx = 0
	}
	return nil
}

// OpenForBackup opens a file or directory, potentially skipping access checks if the backup
// or restore privileges have been acquired.
//
// If the file opened was a directory, it cannot be used with Readdir().
func OpenForBackup(path string, access uint32, share uint32, createmode uint32) (*os.File, error) {
	h, err := fs.CreateFile(path,
		fs.AccessMask(access),
		fs.FileShareMode(share),
		nil,
		fs.FileCreationDisposition(createmode),
		fs.FILE_FLAG_BACKUP_SEMANTICS|fs.FILE_FLAG_OPEN_REPARSE_POINT,
		0)
	if err != nil {
		err = &os.PathError{Op: "open", Path: path, Err: err}
		return nil, err
	}
	return os.NewFile(uintptr(h), path), nil
}
