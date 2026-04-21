package output

import (
	"strings"
	"testing"
)

func TestJSON(t *testing.T) {
	// JSON writes to stdout; we can't easily capture it in a unit test
	// without subprocess trickery, but we can at least verify it doesn't panic.
	JSON(map[string]string{"key": "value"})
}

func TestTable(t *testing.T) {
	// Table writes to stdout; verify it doesn't panic.
	Table([]string{"A", "B"}, [][]string{{"1", "2"}})
}

func TestErrorf(t *testing.T) {
	err := Errorf("something failed: %d", 42)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "something failed: 42") {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestTruncate(t *testing.T) {
	if got := Truncate("hello", 10); got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
	if got := Truncate("hello world", 8); got != "hello..." {
		t.Errorf("expected 'hello...', got %q", got)
	}
}
