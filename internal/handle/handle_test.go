package handle

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Microsoft/go-winio/internal/testutil"
	"golang.org/x/sys/windows"
)

// TODO:

func TestDuplicate(t *testing.T) {
	u := testutil.New(t)
	dir := t.TempDir()

	p := filepath.Join(dir, t.Name()+".txt")
	f, err := os.Create(p)
	u.Must(err)
	defer f.Close()

	b := []byte("hello, test")
	n, err := f.Write(b)
	u.Must(err)
	u.Assert(n != 0, "did not write anything")

	h, err := Duplicate(windows.Handle(f.Fd()))
	u.Must(err)
	ff := os.NewFile(uintptr(h), p)
	defer ff.Close()

	u.Assert(ff != nil, "nil file")
	u.Assert(ff.Fd() != 0, "handle is zero")
	u.Assert(ff.Fd() != uintptr(windows.InvalidHandle), "handle is invalid")
	u.Assert(ff.Fd() != f.Fd(), "handles are not different")
	u.Assert(Compare(windows.Handle(ff.Fd()), windows.Handle(f.Fd())), "handles are different objects")

	_, err = f.Seek(0, 0)
	u.Must(err)

	bb, err := io.ReadAll(ff)
	n = len(bb)
	u.Must(err)
	u.Assert(n != 0, "did not read anything")
	u.Assert(n == len(b), "duplicate file read")
	u.Assert(string(bb) == string(b), fmt.Sprintf("got %s, wanted %s", bb, b))
}
