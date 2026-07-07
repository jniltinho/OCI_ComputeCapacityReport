package output

import "testing"

func TestCenter(t *testing.T) {
	if got := center("abc", 7); got != "  abc  " {
		t.Fatalf("center() = %q", got)
	}
}

func TestColorWrappers(t *testing.T) {
	if Green("ok") == "ok" {
		t.Fatal("Green should add escape codes")
	}
	if Yellow("warn") == "warn" {
		t.Fatal("Yellow should add escape codes")
	}
	if Red("err") == "err" {
		t.Fatal("Red should add escape codes")
	}
}