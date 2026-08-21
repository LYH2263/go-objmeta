package multipart

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"example.com/objmeta/internal/clone"
	"example.com/objmeta/internal/errs"
	"example.com/objmeta/internal/meta"
)

// Manager 分片会话管理。
type Manager struct {
	mu      sync.Mutex
	root    string
	clk     meta.Clock
	maxPart int
	uploads map[string]*session
}

type session struct {
	meta  meta.UploadMeta
	parts map[int]meta.PartMeta
	dir   string
}

func Open(root string, clk meta.Clock, maxPart int) (*Manager, error) {
	dir := filepath.Join(root, "multipart")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	m := &Manager{
		root:    dir,
		clk:     clk,
		maxPart: maxPart,
		uploads: map[string]*session{},
	}
	_ = m.reload()
	return m, nil
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.flushAllLocked()
}

func (m *Manager) Create(ctx context.Context, key, ctype string, um map[string]string) (meta.UploadMeta, error) {
	if err := ctx.Err(); err != nil {
		return meta.UploadMeta{}, errs.WrapCanceled(err)
	}
	id, err := newID()
	if err != nil {
		return meta.UploadMeta{}, err
	}
	dir := filepath.Join(m.root, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return meta.UploadMeta{}, err
	}
	u := meta.UploadMeta{
		ID:          id,
		Key:         key,
		ContentType: ctype,
		UserMeta:    clone.StringMap(um),
		InitiatedAt: m.clk.Now(),
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.uploads[id] = &session{meta: u, parts: map[int]meta.PartMeta{}, dir: dir}
	if err := m.persistLocked(id); err != nil {
		delete(m.uploads, id)
		_ = os.RemoveAll(dir)
		return meta.UploadMeta{}, err
	}
	return u, nil
}

func (m *Manager) PutPart(ctx context.Context, id string, num int, data []byte, sum string) (meta.PartMeta, error) {
	if err := ctx.Err(); err != nil {
		return meta.PartMeta{}, errs.WrapCanceled(err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.uploads[id]
	if !ok {
		return meta.PartMeta{}, errs.ErrUploadNotFound
	}
	path := filepath.Join(s.dir, partName(num))
	if err := writePartFile(ctx, path, data); err != nil {
		return meta.PartMeta{}, err
	}
	pm := meta.PartMeta{Number: num, Size: int64(len(data)), ETag: sum}
	s.parts[num] = pm
	if err := m.persistLocked(id); err != nil {
		return meta.PartMeta{}, err
	}
	return pm, nil
}

func (m *Manager) getLocked(id string) (*session, error) {
	s, ok := m.uploads[id]
	if !ok {
		return nil, errs.ErrUploadNotFound
	}
	return s, nil
}

func newID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func partName(n int) string { return "part-" + hex.EncodeToString([]byte{byte(n >> 8), byte(n)}) }

type diskSession struct {
	Meta  meta.UploadMeta         `json:"meta"`
	Parts map[string]meta.PartMeta `json:"parts"`
}

func (m *Manager) persistLocked(id string) error {
	s := m.uploads[id]
	ds := diskSession{Meta: s.meta, Parts: map[string]meta.PartMeta{}}
	for n, p := range s.parts {
		ds.Parts[itoa(n)] = p
	}
	b, err := json.MarshalIndent(ds, "", "  ")
	if err != nil {
		return errs.WrapPersist("mp-encode", err)
	}
	path := filepath.Join(s.dir, "session.json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return errs.WrapPersist("mp-write", err)
	}
	return os.Rename(tmp, path)
}

func (m *Manager) flushAllLocked() error {
	for id := range m.uploads {
		if err := m.persistLocked(id); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) reload() error {
	ents, err := os.ReadDir(m.root)
	if err != nil {
		return err
	}
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(m.root, e.Name(), "session.json")
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var ds diskSession
		if json.Unmarshal(b, &ds) != nil {
			continue
		}
		s := &session{meta: ds.Meta, parts: map[int]meta.PartMeta{}, dir: filepath.Join(m.root, e.Name())}
		for k, p := range ds.Parts {
			s.parts[atoi(k)] = p
		}
		m.uploads[ds.Meta.ID] = s
	}
	return nil
}

func itoa(n int) string {
	return hex.EncodeToString([]byte{byte(n >> 8), byte(n)})
}

func atoi(s string) int {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) < 2 {
		return 0
	}
	return int(b[0])<<8 | int(b[1])
}

// idle helpers keep API surface thick
func (m *Manager) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.uploads)
}

func (m *Manager) Age(id string) (time.Duration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, err := m.getLocked(id)
	if err != nil {
		return 0, err
	}
	return m.clk.Now().Sub(s.meta.InitiatedAt), nil
}
