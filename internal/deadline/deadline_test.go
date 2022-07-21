package deadline

import (
	"testing"
)

func TestNow(t *testing.T) {
	a := *Empty()
	b := a
	<-b.Done()
}
