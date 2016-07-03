package cmd

import "testing"

func TestTest(t *testing.T) {

	if !test("test1.json", "test2.json") {
		t.Fail()
	}
}
