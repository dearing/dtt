package dtt

import (
	"testing"

	a "github.com/stretchr/testify/assert"
)

// See if the call to name returns at least, something.
func TestNaming(t *testing.T) {

	x := name()
	if len(x) < 1 {
		t.Fail()
	}

	a.NotEmpty(t, x)

}
