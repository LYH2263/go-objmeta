package objmeta_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"example.com/objmeta"
)

func TestBug04_NilUserMetaPutNoPanic(t *testing.T) {
	dir := t.TempDir()
	st, err := objmeta.Open(context.Background(), objmeta.Options{Root: filepath.Join(dir, "o")})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	meta := map[string]string{"a": "1"}
	if _, err := st.Put(context.Background(), "k", bytes.NewReader([]byte("x")), objmeta.PutOptions{UserMeta: meta}); err != nil {
		t.Fatal(err)
	}
	meta["a"] = "MUT"
	info, err := st.Head(context.Background(), "k")
	if err != nil {
		t.Fatal(err)
	}
	if info.UserMeta["a"] != "1" {
		t.Fatalf("user meta aliased: %v", info.UserMeta)
	}
	// nil UserMeta must not panic on Put/Head
	if _, err := st.Put(context.Background(), "k2", bytes.NewReader([]byte("y")), objmeta.PutOptions{UserMeta: nil}); err != nil {
		t.Fatal(err)
	}
	info2, err := st.Head(context.Background(), "k2")
	if err != nil {
		t.Fatal(err)
	}
	if info2.UserMeta != nil && len(info2.UserMeta) != 0 {
		// allow empty map, but must be safe
		_ = info2.UserMeta["missing"]
	}
}
