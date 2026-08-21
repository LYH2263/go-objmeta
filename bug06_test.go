package objmeta_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"example.com/objmeta"
)

func TestBug06_IndexPersistRollback(t *testing.T) {
	// 索引 flush 失败时不得留下数据面孤儿，且内存索引不得脏写。
	dir := t.TempDir()
	root := filepath.Join(dir, "o")
	st, err := objmeta.Open(context.Background(), objmeta.Options{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	idxPath := filepath.Join(root, "meta", "index.json")
	if err := os.Remove(idxPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(idxPath, 0755); err != nil {
		t.Fatal(err)
	}
	_, err = st.Put(context.Background(), "orphan", bytes.NewReader([]byte("payload")), objmeta.PutOptions{})
	if err == nil {
		t.Fatal("expected persist error")
	}
	objPath := filepath.Join(root, "objects", "orphan")
	if _, err := os.Stat(objPath); err == nil {
		t.Fatal("orphan object file left after index persist failure")
	}
	_, err = st.Head(context.Background(), "orphan")
	if err == nil {
		t.Fatal("in-memory index kept dirty entry after persist failure")
	}
}
