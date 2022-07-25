//go:build windows

package file

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Microsoft/go-winio/internal/testutil"
)

// TODO
// - io.ReadAll hangs forever if no deadline set; with deadline fails with "deadline exceeded"

func TestWriteRead(t *testing.T) {
	u := testutil.New(t)
	dir := t.TempDir()

	f, err := OpenFile(filepath.Join(dir, t.Name()+".txt"), GenericRead|GenericWrite, AttributeNormal)
	u.Must(err)
	defer f.Close()

	b := []byte("hello, test")
	n, err := f.Write(b)
	u.Must(err)
	u.Assert(n != 0, "did not write anything")

	bb := make([]byte, len(b)*2)
	n, err = f.Read(bb)
	// f.SetDeadline(time.Now().Add(time.Millisecond))
	// bb, err := io.ReadAll(f)
	// n = len(bb)
	u.Must(err)
	u.Assert(n != 0, "did not read anything")

	bb = bb[:n]
	u.Assert(n == len(b), "read length mismatch")
	u.Assert(string(bb) == string(b), fmt.Sprintf("got %s, wanted %s", bb, b))
}
