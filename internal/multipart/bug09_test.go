package multipart_test

import (
	"context"
	"os"
	"testing"
	"time"

	"example.com/objmeta/internal/clock"
	"example.com/objmeta/internal/multipart"
)

func TestBug09_PartFileSyncBeforeClose(t *testing.T) {
	dir := t.TempDir()
	clk := clock.NewFake(time.Now().UTC())
	m, err := multipart.Open(dir, clk, 100)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	u, err := m.Create(context.Background(), "k", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("part-payload-1234567890")
	if _, err := m.PutPart(context.Background(), u.ID, 1, data, "etag"); err != nil {
		t.Fatal(err)
	}
	path, err := m.PartPath(u.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Fatalf("part file content=%q want %q (Close before Sync?)", got, data)
	}
}
