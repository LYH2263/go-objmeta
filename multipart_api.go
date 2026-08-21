package objmeta

import (
	"context"
	"io"

	"example.com/objmeta/internal/clone"
	"example.com/objmeta/internal/errs"
	"example.com/objmeta/internal/etag"
	"example.com/objmeta/internal/meta"
	"example.com/objmeta/internal/store"
	"example.com/objmeta/internal/validate"
)

// CreateMultipart 初始化分片上传会话。
func (s *Store) CreateMultipart(ctx context.Context, key string, opt MultipartOptions) (MultipartUpload, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.guard(); err != nil {
		return MultipartUpload{}, err
	}
	if err := validate.Key(key); err != nil {
		return MultipartUpload{}, err
	}
	if err := ctx.Err(); err != nil {
		return MultipartUpload{}, errs.WrapCanceled(err)
	}
	um := clone.StringMap(opt.UserMeta)
	u, err := s.mp.Create(ctx, key, opt.ContentType, um)
	if err != nil {
		return MultipartUpload{}, err
	}
	s.met.IncMultipartCreate()
	return MultipartUpload{UploadID: u.ID, Key: u.Key, InitiatedAt: u.InitiatedAt}, nil
}

// UploadPart 上传一个分片；尊重 ctx 取消。
func (s *Store) UploadPart(ctx context.Context, uploadID string, partNumber int, body io.Reader) (PartInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.guard(); err != nil {
		return PartInfo{}, err
	}
	if partNumber < 1 || partNumber > s.maxParts {
		return PartInfo{}, ErrInvalidPart
	}
	if err := ctx.Err(); err != nil {
		return PartInfo{}, errs.WrapCanceled(err)
	}
	data, err := store.ReadAllLimited(ctx, body, s.maxBytes)
	if err != nil {
		return PartInfo{}, err
	}
	owned := clone.Bytes(data)
	sum := etag.Of(owned)
	pi, err := s.mp.PutPart(ctx, uploadID, partNumber, owned, sum)
	if err != nil {
		return PartInfo{}, err
	}
	s.met.IncPart()
	return PartInfo{PartNumber: pi.Number, Size: pi.Size, ETag: pi.ETag}, nil
}

// CompleteMultipart 按序合并分片；缺片时返回可 errors.Is 的 ErrIncompleteParts。
func (s *Store) CompleteMultipart(ctx context.Context, uploadID string, parts []int) (ObjectInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.guard(); err != nil {
		return ObjectInfo{}, err
	}
	if err := ctx.Err(); err != nil {
		return ObjectInfo{}, errs.WrapCanceled(err)
	}
	partsCopy := clone.Ints(parts)
	u, merged, sum, err := s.mp.Complete(ctx, uploadID, partsCopy)
	if err != nil {
		return ObjectInfo{}, err
	}
	m := meta.ObjectMeta{
		Key:          u.Key,
		Size:         int64(len(merged)),
		ETag:         sum,
		ContentType:  u.ContentType,
		UserMeta:     clone.StringMap(u.UserMeta),
		LastModified: s.clk.Now(),
		Version:      1,
	}
	if exist, ok := s.index.Get(u.Key); ok {
		m.Version = exist.Version + 1
	}
	if err := s.backend.WriteObject(ctx, u.Key, merged); err != nil {
		return ObjectInfo{}, err
	}
	if err := s.index.Put(m); err != nil {
		_ = s.backend.RemoveObject(u.Key)
		return ObjectInfo{}, err
	}
	s.met.IncMultipartComplete()
	return toPublicInfo(m), nil
}

// AbortMultipart 中止会话；持久化失败时不得留下半清状态。
func (s *Store) AbortMultipart(ctx context.Context, uploadID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.guard(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return errs.WrapCanceled(err)
	}
	if err := s.mp.Abort(ctx, uploadID); err != nil {
		return err
	}
	s.met.IncMultipartAbort()
	return nil
}

// ListParts 列出会话已上传分片。
func (s *Store) ListParts(ctx context.Context, uploadID string) ([]PartInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := s.guard(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, errs.WrapCanceled(err)
	}
	ps, err := s.mp.ListParts(uploadID)
	if err != nil {
		return nil, err
	}
	out := make([]PartInfo, 0, len(ps))
	for _, p := range ps {
		out = append(out, PartInfo{PartNumber: p.Number, Size: p.Size, ETag: p.ETag})
	}
	return out, nil
}
