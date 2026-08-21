package objmeta

import (
	"bytes"
	"context"
	"io"
	"sync"

	"example.com/objmeta/internal/clock"
	"example.com/objmeta/internal/clone"
	"example.com/objmeta/internal/errs"
	"example.com/objmeta/internal/etag"
	"example.com/objmeta/internal/fsbackend"
	"example.com/objmeta/internal/list"
	"example.com/objmeta/internal/meta"
	"example.com/objmeta/internal/metrics"
	"example.com/objmeta/internal/multipart"
	"example.com/objmeta/internal/persist"
	"example.com/objmeta/internal/policy"
	"example.com/objmeta/internal/store"
	"example.com/objmeta/internal/validate"
)

// Store 对象元数据门面；线程安全。
type Store struct {
	mu       sync.RWMutex
	closed   bool
	root     string
	backend  *fsbackend.Backend
	index    *persist.Index
	mp       *multipart.Manager
	pol      *policy.Guard
	clk      meta.Clock
	met      *metrics.Registry
	maxBytes int64
	maxParts int
	hot      map[string][]byte // Put 热缓存；必须与调用方切片隔离
	listEng  *list.Engine      // 列举 CommonPrefixes 复用缓冲
}

// Open 打开或创建本地对象库。
func Open(ctx context.Context, opt Options) (*Store, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if opt.Root == "" {
		return nil, errs.ErrInvalidKey
	}
	clk := opt.Clock
	if clk == nil {
		clk = clock.System()
	}
	maxB := opt.MaxObjectBytes
	if maxB <= 0 {
		maxB = 64 << 20
	}
	maxP := opt.MaxParts
	if maxP <= 0 {
		maxP = 10000
	}
	be, err := fsbackend.Open(opt.Root)
	if err != nil {
		return nil, err
	}
	idx, err := persist.OpenIndex(opt.Root)
	if err != nil {
		_ = be.Close()
		return nil, err
	}
	mp, err := multipart.Open(opt.Root, clk, maxP)
	if err != nil {
		_ = idx.Close()
		_ = be.Close()
		return nil, err
	}
	st := &Store{
		root:     opt.Root,
		backend:  be,
		index:    idx,
		mp:       mp,
		pol:      policy.New(),
		clk:      clk,
		met:      metrics.New(),
		maxBytes: maxB,
		maxParts: maxP,
		hot:      make(map[string][]byte),
		listEng:  &list.Engine{},
	}
	return st, nil
}

// Close 关闭底层句柄；之后写接口返回 ErrClosed。
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	var first error
	if err := s.mp.Close(); err != nil && first == nil {
		first = err
	}
	if err := s.index.Close(); err != nil && first == nil {
		first = err
	}
	if err := s.backend.Close(); err != nil && first == nil {
		first = err
	}
	// 保持 backend/index 非 nil：写路径先查 closed，避免 Close 后 nil 解引用。
	return first
}

func (s *Store) guard() error {
	if s.closed {
		return ErrClosed
	}
	if s.backend == nil || s.index == nil {
		return ErrClosed
	}
	return nil
}

// Put 写入对象；body 在入口拷贝，调用方事后改缓冲不影响已存内容。
func (s *Store) Put(ctx context.Context, key string, body io.Reader, opt PutOptions) (ObjectInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.guard(); err != nil {
		return ObjectInfo{}, err
	}
	if err := validate.Key(key); err != nil {
		return ObjectInfo{}, err
	}
	if err := ctx.Err(); err != nil {
		return ObjectInfo{}, errs.WrapCanceled(err)
	}
	data, err := store.ReadAllLimited(ctx, body, s.maxBytes)
	if err != nil {
		return ObjectInfo{}, err
	}
	return s.putOwnedLocked(ctx, key, clone.Bytes(data), opt)
}

// PutBytes 写入调用方 []byte；必须深拷贝后再入库。
func (s *Store) PutBytes(ctx context.Context, key string, data []byte, opt PutOptions) (ObjectInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.guard(); err != nil {
		return ObjectInfo{}, err
	}
	if err := validate.Key(key); err != nil {
		return ObjectInfo{}, err
	}
	if err := ctx.Err(); err != nil {
		return ObjectInfo{}, errs.WrapCanceled(err)
	}
	owned := clone.Bytes(data)
	return s.putOwnedLocked(ctx, key, owned, opt)
}

func (s *Store) putOwnedLocked(ctx context.Context, key string, owned []byte, opt PutOptions) (ObjectInfo, error) {
	sum := etag.Of(owned)
	exist, ok := s.index.Get(key)
	if err := s.pol.CheckPut(ok, exist.ETag, opt.IfNoneMatch, opt.IfMatchETag); err != nil {
		return ObjectInfo{}, err
	}
	um := clone.StringMap(opt.UserMeta)
	m := meta.ObjectMeta{
		Key:          key,
		Size:         int64(len(owned)),
		ETag:         sum,
		ContentType:  opt.ContentType,
		UserMeta:     um,
		LastModified: s.clk.Now(),
		Version:      exist.Version + 1,
	}
	if err := s.backend.WriteObject(ctx, key, owned); err != nil {
		return ObjectInfo{}, err
	}
	if err := s.index.Put(m); err != nil {
		_ = s.backend.RemoveObject(key)
		return ObjectInfo{}, err
	}
	s.hot[key] = owned
	s.met.IncPut()
	return toPublicInfo(m), nil
}

// Get 读取对象；返回的 Info 与内部索引隔离。
func (s *Store) Get(ctx context.Context, key string) (*Object, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := s.guard(); err != nil {
		return nil, err
	}
	if err := validate.Key(key); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, errs.WrapCanceled(err)
	}
	m, ok := s.index.Get(key)
	if !ok {
		return nil, ErrNotFound
	}
	if hot, ok := s.hot[key]; ok {
		s.met.IncGet()
		return &Object{Info: toPublicInfo(m), Body: io.NopCloser(bytes.NewReader(hot))}, nil
	}
	rc, err := s.backend.OpenObject(ctx, key)
	if err != nil {
		return nil, err
	}
	s.met.IncGet()
	return &Object{Info: toPublicInfo(m), Body: rc}, nil
}

// Head 仅返回元数据。
func (s *Store) Head(ctx context.Context, key string) (ObjectInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := s.guard(); err != nil {
		return ObjectInfo{}, err
	}
	if err := validate.Key(key); err != nil {
		return ObjectInfo{}, err
	}
	if err := ctx.Err(); err != nil {
		return ObjectInfo{}, errs.WrapCanceled(err)
	}
	m, ok := s.index.Get(key)
	if !ok {
		return ObjectInfo{}, ErrNotFound
	}
	return toPublicInfo(m), nil
}

// Delete 删除对象。
func (s *Store) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.guard(); err != nil {
		return err
	}
	if err := validate.Key(key); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return errs.WrapCanceled(err)
	}
	if _, ok := s.index.Get(key); !ok {
		return ErrNotFound
	}
	if err := s.index.Delete(key); err != nil {
		return err
	}
	delete(s.hot, key)
	_ = s.backend.RemoveObject(key)
	s.met.IncDelete()
	return nil
}

// List 前缀列举；结果切片为拷贝。
func (s *Store) List(ctx context.Context, opt ListOptions) (ListResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := s.guard(); err != nil {
		return ListResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return ListResult{}, errs.WrapCanceled(err)
	}
	all := s.index.Snapshot()
	raw := s.listEng.Filter(all, list.Options{
		Prefix:    opt.Prefix,
		Delimiter: opt.Delimiter,
		Limit:     opt.Limit,
		Cursor:    opt.Cursor,
	})
	entries := make([]ObjectInfo, 0, len(raw.Entries))
	for _, m := range raw.Entries {
		entries = append(entries, toPublicInfo(m))
	}
	// 拷贝 common prefixes，避免调用方改写内部缓冲
	prefs := clone.Strings(raw.CommonPrefixes)
	return ListResult{
		Entries:     entries,
		CommonPrefs: prefs,
		NextCursor:  raw.NextCursor,
		Truncated:   raw.Truncated,
	}, nil
}

// Stats 返回计数器快照。
func (s *Store) Stats() metrics.Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.met.Snapshot()
}
