//go:build windows

package security

import (
	"errors"
	"testing"

	"golang.org/x/sys/windows"
)

func TestLookupInvalidSid(t *testing.T) {
	_, err := LookupSIDByName(".\\weoifjdsklfj")
	var aerr *AccountLookupError
	if !errors.As(err, &aerr) || !errors.Is(err, windows.ERROR_NONE_MAPPED) {
		t.Fatalf("expected AccountLookupError with ERROR_NONE_MAPPED, got %s", err)
	}
}

func TestLookupInvalidName(t *testing.T) {
	_, err := LookupNameBySID("notasid")
	var aerr *AccountLookupError
	if !errors.As(err, &aerr) || !errors.Is(aerr.Err, windows.ERROR_INVALID_SID) {
		t.Fatalf("expected AccountLookupError with ERROR_INVALID_SID got %s", err)
	}
}

func TestLookupValidSid(t *testing.T) {
	everyone := "S-1-1-0"
	name, err := LookupNameBySID(everyone)
	if err != nil {
		t.Fatalf("expected a valid account name, got %v", err)
	}

	sid, err := LookupSIDByName(name)
	if err != nil || sid != everyone {
		t.Fatalf("expected %s, got %s, %s", everyone, sid, err)
	}
}

func TestLookupEmptyNameFails(t *testing.T) {
	_, err := LookupSIDByName("")
	var aerr *AccountLookupError
	if !errors.As(err, &aerr) || !errors.Is(aerr.Err, windows.ERROR_NONE_MAPPED) {
		t.Fatalf("expected AccountLookupError with ERROR_NONE_MAPPED, got %s", err)
	}
}
