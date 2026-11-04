package main

import "testing"

func TestGreeting(t *testing.T) {
	want := "chain-reader-cli"
	if got := greeting(); got != want {
		t.Errorf("greeting() = %q, want %q", got, want)
	}
}
