package winio

import (
	"github.com/Microsoft/go-winio/pkg/fs"
)

// ExtendedAttribute represents a single Windows EA.
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/fs.ExtendedAttribute instead.
type ExtendedAttribute = fs.ExtendedAttribute

// DecodeExtendedAttributes decodes a list of EAs from a FILE_FULL_EA_INFORMATION
// buffer retrieved from BackupRead, ZwQueryEaFile, etc.
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/fs.DecodeExtendedAttributes instead.
func DecodeExtendedAttributes(b []byte) (eas []ExtendedAttribute, err error) {
	return fs.DecodeExtendedAttributes(b)
}

// EncodeExtendedAttributes encodes a list of EAs into a FILE_FULL_EA_INFORMATION
// buffer for use with BackupWrite, ZwSetEaFile, etc.
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/fs.EncodeExtendedAttributes instead.
func EncodeExtendedAttributes(eas []ExtendedAttribute) ([]byte, error) {
	return fs.EncodeExtendedAttributes(eas)
}
