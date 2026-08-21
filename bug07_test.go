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

type slowReader struct {
	n int
}

func (s *slowReader) Read(p []byte) (int, error) {
	if s.n > 0 {
		time.Sleep(30 * time.Millisecond)
		p[0] = 'a'
		s.n--
		return 1, nil
	}
	return 0, io.EOF
}

func TestBug07_PutHonorsContextCancel(t *testing.T) {
	dir := t.TempDir()
	st, err := objmeta.Open(context.Background(), objmeta.Options{Root: filepath.Join(dir, "o")})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx, cancel := context.WithCancel(context.Background())
	r := &slowReader{n: 40}
	go func() {
		time.Sleep(40 * time.Millisecond)
		cancel()
	}()
	_, err = st.Put(ctx, "k", r, objmeta.PutOptions{})
	if err == nil {
		t.Fatal("expected cancel error")
	}
	if !errors.Is(err, objmeta.ErrCanceled) && !errors.Is(err, context.Canceled) {
		t.Fatalf("want canceled, got %v", err)
	}
}
