package operations

import (
	"testing"
)

func TestHttpRequest(t *testing.T) {
	t.Skip("Skipping testing in CI environment")
	actual, error := SendStringToPastebin("test string")
	if error != nil {
		t.Fatalf("expected no error, got %v", error)
	}
	if actual == "" {
		t.Fatalf("expected not empty, got %v", actual)
	}
}
