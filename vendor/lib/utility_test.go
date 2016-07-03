package lib

import (
	"testing"
)

func TestNaming(t *testing.T) {

	x := name()
	if len(x) < 1 {
		t.Fail()
	}
}
