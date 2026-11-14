package model

import "testing"

func TestReads(t *testing.T) {
	if Oldest > Version {
		t.Fatalf("oldest_model_version %d is newer than model_version %d", Oldest, Version)
	}
	for in, want := range map[string]bool{"0": true, "1": true, "2": false, "12": false, "99999999999999999999": false, "": false, "-1": false} {
		if got := Reads(in); got != want {
			t.Errorf("%q: %v, want %v", in, got, want)
		}
	}
}

func TestRange(t *testing.T) {
	if got := Range(); got != "model versions 0 to 1" {
		t.Errorf("got %q", got)
	}
}
