package objmeta_test

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	"example.com/objmeta"
)

func TestBug03_PutAfterCloseNoPanic(t *testing.T) {
	dir := t.TempDir()
	st, err := objmeta.Open(context.Background(), objmeta.Options{Root: filepath.Join(dir, "o")})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("Put after Close panicked: %v", rec)
		}
	}()
	_, err = st.Put(context.Background(), "k", bytes.NewReader([]byte("x")), objmeta.PutOptions{})
	if !errors.Is(err, objmeta.ErrClosed) {
		t.Fatalf("want ErrClosed, got %v", err)
	}
}
