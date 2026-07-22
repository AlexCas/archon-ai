package mapgen

import (
	"errors"
	"strings"
	"testing"
)

func TestSplice_AbsentMarkers(t *testing.T) {
	existing := "# Preamble\n\nAuthored prose.\n"
	got, err := Splice(existing, "body\n")
	if err != nil {
		t.Fatalf("Splice() error = %v", err)
	}
	want := "# Preamble\n\nAuthored prose.\n" + mapStart + "\nbody\n" + mapEnd + "\n"
	if got != want {
		t.Errorf("Splice() = %q, want %q", got, want)
	}
}

func TestSplice_ExistingMarkers_ReplacesBody(t *testing.T) {
	existing := "before\n" + mapStart + "\nold body\n" + mapEnd + "\nafter\n"
	got, err := Splice(existing, "new body\n")
	if err != nil {
		t.Fatalf("Splice() error = %v", err)
	}
	want := "before\n" + mapStart + "\nnew body\n" + mapEnd + "\nafter\n"
	if got != want {
		t.Errorf("Splice() = %q, want %q", got, want)
	}
}

func TestSplice_PreservesProseOutsideMarkers(t *testing.T) {
	existing := "AUTHORED BEFORE\n" + mapStart + "\nold\n" + mapEnd + "\nAUTHORED AFTER\n"
	got, err := Splice(existing, "new\n")
	if err != nil {
		t.Fatalf("Splice() error = %v", err)
	}
	if !strings.Contains(got, "AUTHORED BEFORE") || !strings.Contains(got, "AUTHORED AFTER") {
		t.Errorf("Splice() dropped authored prose; got %q", got)
	}
}

func TestSplice_NestedRegion(t *testing.T) {
	existing := mapStart + "\na\n" + mapEnd + "\n" + mapStart + "\nb\n" + mapEnd + "\n"
	_, err := Splice(existing, "body\n")
	if !errors.Is(err, ErrNestedRegion) {
		t.Errorf("Splice() error = %v, want ErrNestedRegion", err)
	}
}
