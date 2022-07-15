package exporters

import (
	"unsafe"

	"github.com/Microsoft/go-winio/pkg/guid"
)

type SpanID = [8]byte

func SpanIDToGUID(id SpanID) (g guid.GUID) {
	// casting directly would read off the end of the SpanID array
	p := (*SpanID)(unsafe.Pointer(&g))
	*p = id
	return g
}

type TraceID = [16]byte

func TraceIDToGUID(id TraceID) guid.GUID {
	g := *((*guid.GUID)(unsafe.Pointer(&id[0])))
	return g
}
