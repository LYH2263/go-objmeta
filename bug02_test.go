package objmeta_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"example.com/objmeta"
)

func TestBug02_ListPrefixSliceAlias(t *testing.T) {
	dir := t.TempDir()
	st, err := objmeta.Open(context.Background(), objmeta.Options{Root: filepath.Join(dir, "o")})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	for _, k := range []string{"p/a", "p/b", "p/c/x"} {
		if _, err := st.Put(context.Background(), k, bytes.NewReader([]byte("v")), objmeta.PutOptions{}); err != nil {
			t.Fatal(err)
		}
	}
	res, err := st.List(context.Background(), objmeta.ListOptions{Prefix: "p/", Delimiter: "/", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.CommonPrefs) == 0 {
		t.Fatal("expected common prefix")
	}
	res2, err := st.List(context.Background(), objmeta.ListOptions{Prefix: "p/", Delimiter: "/", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	orig := append([]string(nil), res2.CommonPrefs...)
	res.CommonPrefs[0] = "MUTATED/"
	if res2.CommonPrefs[0] != orig[0] {
		t.Fatalf("list common prefix aliased: got %q want %q", res2.CommonPrefs[0], orig[0])
	}
}
