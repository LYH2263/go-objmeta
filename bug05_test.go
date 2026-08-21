package objmeta_test

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	"example.com/objmeta"
)

func TestBug05_CompleteMissingPartsWraps(t *testing.T) {
	dir := t.TempDir()
	st, err := objmeta.Open(context.Background(), objmeta.Options{Root: filepath.Join(dir, "o")})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	up, err := st.CreateMultipart(context.Background(), "big", objmeta.MultipartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.UploadPart(context.Background(), up.UploadID, 1, bytes.NewReader([]byte("aa"))); err != nil {
		t.Fatal(err)
	}
	_, err = st.CompleteMultipart(context.Background(), up.UploadID, []int{1, 2})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, objmeta.ErrIncompleteParts) {
		t.Fatalf("want ErrIncompleteParts wrap, got %v", err)
	}
}
