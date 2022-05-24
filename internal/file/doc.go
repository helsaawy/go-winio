// This package exposes common Windows file APIs.
package file

//go:generate go run github.com/Microsoft/go-winio/tools/mkwinsyscall -import github.com/Microsoft/go-winio/internal/wintype -output zfile.go file.go
