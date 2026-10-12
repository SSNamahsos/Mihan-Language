package vm

import (
	"testing"
	"unsafe"
)

func TestAValueFitsInSixWords(t *testing.T) {
	const want = 48
	if got := unsafe.Sizeof(Value{}); got != want {
		t.Errorf("a value is %d bytes, not %d.\n"+
			"Every push and every pop copies all of it, so growing it slows down every "+
			"arithmetic program in the language. If a new field is genuinely needed, "+
			"change this number too, so the cost is a decision rather than a surprise.",
			got, want)
	}
}

func TestAFrameIsSmall(t *testing.T) {
	if got := unsafe.Sizeof(frame{}); got > 80 {
		t.Errorf("a frame is %d bytes; every call makes one, and a recursive function "+
			"makes one per turn", got)
	}
}
