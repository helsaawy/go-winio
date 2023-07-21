//go:build windows

package fs

import (
	"os"
	"reflect"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	testEAs = []ExtendedAttribute{
		{Name: "foo", Value: []byte("bar")},
		{Name: "fizz", Value: []byte("buzz")},
	}

	testEAsEncoded = []byte{16, 0, 0, 0, 0, 3, 3, 0, 102, 111, 111, 0, 98, 97, 114, 0, 0,
		0, 0, 0, 0, 4, 4, 0, 102, 105, 122, 122, 0, 98, 117, 122, 122, 0, 0, 0}
	testEAsNotPadded = testEAsEncoded[0 : len(testEAsEncoded)-3]
	testEAsTruncated = testEAsEncoded[0:20]
)

func Test_RoundTripEAs(t *testing.T) {
	b, err := EncodeExtendedAttributes(testEAs)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testEAsEncoded, b) {
		t.Fatalf("encoded mismatch %v %v", testEAsEncoded, b)
	}
	eas, err := DecodeExtendedAttributes(b)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testEAs, eas) {
		t.Fatalf("mismatch %+v %+v", testEAs, eas)
	}
}

func Test_EAsDontNeedPaddingAtEnd(t *testing.T) {
	eas, err := DecodeExtendedAttributes(testEAsNotPadded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testEAs, eas) {
		t.Fatalf("mismatch %+v %+v", testEAs, eas)
	}
}

func Test_TruncatedEAsFailCorrectly(t *testing.T) {
	_, err := DecodeExtendedAttributes(testEAsTruncated)
	if err == nil {
		t.Fatal("expected error")
	}
}

func Test_NilEAsEncodeAndDecodeAsNil(t *testing.T) {
	b, err := EncodeExtendedAttributes(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 0 {
		t.Fatal("expected empty")
	}
	eas, err := DecodeExtendedAttributes(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(eas) != 0 {
		t.Fatal("expected empty")
	}
}

// Test_SetFileEA makes sure that the test buffer is actually parsable by NtSetEaFile.
func Test_SetFileEA(t *testing.T) {
	f, err := os.CreateTemp("", "winio")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	ntdll := windows.MustLoadDLL("ntdll.dll")
	ntSetEaFile := ntdll.MustFindProc("NtSetEaFile")
	var iosb [2]uintptr
	r, _, _ := ntSetEaFile.Call(f.Fd(),
		uintptr(unsafe.Pointer(&iosb[0])),
		uintptr(unsafe.Pointer(&testEAsEncoded[0])),
		uintptr(len(testEAsEncoded)))
	if r != 0 {
		t.Fatalf("NtSetEaFile failed with %08x", r)
	}
}
