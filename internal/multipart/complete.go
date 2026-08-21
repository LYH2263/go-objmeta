package multipart

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"example.com/objmeta/internal/errs"
	"example.com/objmeta/internal/etag"
	"example.com/objmeta/internal/meta"
)

// Complete 校验分片齐全后合并；缺片返回 WrapIncomplete。
func (m *Manager) Complete(ctx context.Context, id string, parts []int) (meta.UploadMeta, []byte, string, error) {
	if err := ctx.Err(); err != nil {
		return meta.UploadMeta{}, nil, "", errs.WrapCanceled(err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	s, err := m.getLocked(id)
	if err != nil {
		return meta.UploadMeta{}, nil, "", err
	}
	var missing []int
	for _, n := range parts {
		if _, ok := s.parts[n]; !ok {
			missing = append(missing, n)
		}
	}
	if len(missing) > 0 {

		return meta.UploadMeta{}, nil, "", fmt.Errorf("missing parts %v", missing)
	}
	var merged []byte
	var etags []string
	for _, n := range parts {
		if err := ctx.Err(); err != nil {
			return meta.UploadMeta{}, nil, "", errs.WrapCanceled(err)
		}
		path := filepath.Join(s.dir, partName(n))
		b, err := os.ReadFile(path)
		if err != nil {
			return meta.UploadMeta{}, nil, "", err
		}
		merged = append(merged, b...)
		etags = append(etags, s.parts[n].ETag)
	}
	sum := etag.Multipart(etags)
	u := s.meta
	delete(m.uploads, id)
	_ = os.RemoveAll(s.dir)
	return u, merged, sum, nil
}
