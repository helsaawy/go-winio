//go:build windows

package winio

import (
	"syscall"

	"github.com/Microsoft/go-winio/pkg/security"
	"golang.org/x/sys/windows"
)

const (
	// Deprecated: use golang.org/x/sys/windows.SE_PRIVILEGE_ENABLED instead.
	//revive:disable-next-line:var-naming ALL_CAPS
	SE_PRIVILEGE_ENABLED = windows.SE_PRIVILEGE_ENABLED

	// Deprecated: use golang.org/x/sys/windows.ERROR_NOT_ALL_ASSIGNED instead.
	//revive:disable-next-line:var-naming ALL_CAPS
	ERROR_NOT_ALL_ASSIGNED syscall.Errno = windows.ERROR_NOT_ALL_ASSIGNED
)

// Deprecated: use constants in github.com/Microsoft/go-winio/pkg/security instead.
const (
	SeBackupPrivilege   = "SeBackupPrivilege"
	SeRestorePrivilege  = "SeRestorePrivilege"
	SeSecurityPrivilege = "SeSecurityPrivilege"
)

// PrivilegeError represents an error enabling privileges.
type PrivilegeError = security.PrivilegeError

// RunWithPrivilege enables a single privilege for a function call.
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/security.RunWithPrivilege instead.
func RunWithPrivilege(name string, fn func() error) error {
	return security.RunWithPrivilege(name, fn)
}

// RunWithPrivileges enables privileges for a function call.
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/security.RunWithPrivileges instead.
func RunWithPrivileges(names []string, fn func() error) error {
	return security.RunWithPrivileges(names, fn)
}

// Deprecated: use github.com/Microsoft/go-winio/pkg/security.EnableProcessPrivileges instead.
func EnableProcessPrivileges(names []string) error {
	return security.EnableProcessPrivileges(names)
}

// DisableProcessPrivileges disables privileges globally for the process.
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/security.DisableProcessPrivileges instead.
func DisableProcessPrivileges(names []string) error {
	return security.DisableProcessPrivileges(names)
}
