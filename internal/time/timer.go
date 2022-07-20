package time

import (
	_ "unsafe" // for go:linkname

	"time"
)

//todo: remove in go1.18
//nolint:predeclared
type any = interface{}

// copied directly from "time" sleep.go
// must be kept up to date with ../runtime/time.go:/^type timer
//nolint:structcheck
type timer struct {
	pp       uintptr
	when     int64
	period   int64
	f        func(any, uintptr)
	arg      any
	seq      uintptr
	nextwhen int64
	status   uint32
}

//go:linkname startTimer time.startTimer
func startTimer(*timer)

//go:linkname resetTimer time.resetTimer
func resetTimer(*timer, int64) bool

//go:linkname when time.when
func when(d time.Duration) int64
