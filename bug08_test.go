package objmeta_test

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"testing"
	"time"

	"example.com/objmeta"
)

type slowPartReader struct {
	left int
}

func (s *slowPartReader) Read(p []byte) (int, error) {
	if s.left <= 0 {
		return 0, io.EOF
	}
	time.Sleep(25 * time.Millisecond)
	p[0] = 'b'
	s.left--
	return 1, nil
}

func TestBug08_UploadPartHonorsContextCancel(t *testing.T) {
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
	ctx, cancel := context.WithCancel(context.Background())
	r := &slowPartReader{left: 40}
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	_, err = st.UploadPart(ctx, up.UploadID, 1, r)
	if err == nil {
		t.Fatal("expected cancel")
	}
	if !errors.Is(err, objmeta.ErrCanceled) && !errors.Is(err, context.Canceled) {
		t.Fatalf("want canceled, got %v", err)
	}
}
