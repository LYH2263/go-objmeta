package objmeta_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"example.com/objmeta"
)

func TestPutGetRoundTrip(t *testing.T) {
	dir := t.TempDir()
	st, err := objmeta.Open(context.Background(), objmeta.Options{Root: filepath.Join(dir, "o")})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	info, err := st.Put(context.Background(), "a/b.txt", bytes.NewReader([]byte("hello")), objmeta.PutOptions{ContentType: "text/plain"})
	if err != nil {
		t.Fatal(err)
	}
	if info.Size != 5 {
		t.Fatalf("size=%d", info.Size)
	}
	obj, err := st.Get(context.Background(), "a/b.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer obj.Body.Close()
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(obj.Body); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "hello" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestMultipartHappyPath(t *testing.T) {
	dir := t.TempDir()
	st, err := objmeta.Open(context.Background(), objmeta.Options{Root: filepath.Join(dir, "o")})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	up, err := st.CreateMultipart(context.Background(), "big.bin", objmeta.MultipartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.UploadPart(context.Background(), up.UploadID, 1, bytes.NewReader([]byte("aa"))); err != nil {
		t.Fatal(err)
	}
	if _, err := st.UploadPart(context.Background(), up.UploadID, 2, bytes.NewReader([]byte("bb"))); err != nil {
		t.Fatal(err)
	}
	info, err := st.CompleteMultipart(context.Background(), up.UploadID, []int{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if info.Size != 4 {
		t.Fatalf("size=%d", info.Size)
	}
}
