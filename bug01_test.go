package objmeta_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"example.com/objmeta"
)

func TestBug01_PutBodySliceAlias(t *testing.T) {
	dir := t.TempDir()
	st, err := objmeta.Open(context.Background(), objmeta.Options{Root: filepath.Join(dir, "o")})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	buf := []byte("abcdef")
	if _, err := st.PutBytes(context.Background(), "k1", buf, objmeta.PutOptions{}); err != nil {
		t.Fatal(err)
	}
	buf[0] = 'X'
	obj, err := st.Get(context.Background(), "k1")
	if err != nil {
		t.Fatal(err)
	}
	defer obj.Body.Close()
	got := new(bytes.Buffer)
	if _, err := got.ReadFrom(obj.Body); err != nil {
		t.Fatal(err)
	}
	if got.String() != "abcdef" {
		t.Fatalf("stored body corrupted by caller alias: %q", got.String())
	}
}
