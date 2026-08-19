package tests

import "testing"

func TestSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	if 1+1 != 2 {
		t.Fatal("smoke arithmetic failed")
	}
}
