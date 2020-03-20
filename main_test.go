package main

import "testing"

func TestDummy(t *testing.T) {
	res := Dummy()
	if res != 1 {
		t.Errorf("Bad dummy")
	}
}