package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot resolve home dir")
	}

	got := expandPath("~/test/config")
	want := filepath.Join(home, "test/config")
	if got != want {
		t.Fatalf("expandPath() = %q, want %q", got, want)
	}

	if expandPath("/etc/oci/config") != "/etc/oci/config" {
		t.Fatal("absolute path should remain unchanged")
	}
}