package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"example.com/objmeta/internal/errs"
	"example.com/objmeta/internal/meta"
)

// Index 内存+磁盘元数据索引。
type Index struct {
	mu   sync.RWMutex
	path string
	data map[string]meta.ObjectMeta
}

// OpenIndex 加载或创建索引。
func OpenIndex(root string) (*Index, error) {
	path := filepath.Join(root, "meta", "index.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, errs.WrapPersist("mkdir", err)
	}
	idx := &Index{path: path, data: map[string]meta.ObjectMeta{}}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return idx, idx.flushLocked()
		}
		return nil, errs.WrapPersist("read", err)
	}
	if len(b) > 0 {
		if err := json.Unmarshal(b, &idx.data); err != nil {
			return nil, errs.WrapPersist("decode", err)
		}
	}
	return idx, nil
}

func (idx *Index) Close() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	return idx.flushLocked()
}

func (idx *Index) Get(key string) (meta.ObjectMeta, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	m, ok := idx.data[key]
	if !ok {
		return meta.ObjectMeta{}, false
	}
	return meta.CloneMeta(m), true
}

func (idx *Index) Put(m meta.ObjectMeta) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	prev, had := idx.data[m.Key]
	idx.data[m.Key] = meta.CloneMeta(m)
	if err := idx.flushLocked(); err != nil {
		if had {
			idx.data[m.Key] = prev
		} else {
			delete(idx.data, m.Key)
		}
		return err
	}
	return nil
}

func (idx *Index) Delete(key string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	delete(idx.data, key)
	return idx.flushLocked()
}

func (idx *Index) Snapshot() []meta.ObjectMeta {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	out := make([]meta.ObjectMeta, 0, len(idx.data))
	for _, m := range idx.data {
		out = append(out, meta.CloneMeta(m))
	}
	return out
}

func (idx *Index) flushLocked() error {
	b, err := json.MarshalIndent(idx.data, "", "  ")
	if err != nil {
		return errs.WrapPersist("encode", err)
	}
	tmp := idx.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return errs.WrapPersist("write", err)
	}
	if err := os.Rename(tmp, idx.path); err != nil {
		return errs.WrapPersist("rename", err)
	}
	return nil
}
