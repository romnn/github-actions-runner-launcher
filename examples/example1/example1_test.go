package main

import (
	"testing"
)

func TestCli(t *testing.T) {
	out := run()
	expected := "t.b.a. Contributions are welcome :)"
	if out != expected {
		t.Errorf("Got %s but expected %s", out, expected)
	}
}
