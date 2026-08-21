package fsbackend_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"example.com/objmeta/internal/fsbackend"
)

func TestBug10_ObjectWriteSyncBeforeClose(t *testing.T) {
	dir := t.TempDir()
	b, err := fsbackend.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	data := []byte("object-bytes-abcdefghij")
	if err := b.WriteObject(context.Background(), "a/b", data); err != nil {
		// Sync after Close returns error on Windows/Linux — that is the bug surface
		t.Fatalf("WriteObject failed (likely Sync after Close): %v", err)
	}
	path := filepath.Join(dir, "objects", "a", "b")
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Fatalf("got %q", got)
	}
}
