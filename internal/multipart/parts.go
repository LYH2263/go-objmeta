package multipart

import (
	"context"
	"os"
	"path/filepath"
	"sort"

	"example.com/objmeta/internal/errs"
	"example.com/objmeta/internal/meta"
)

// ListParts 返回已上传分片（按号排序拷贝）。
func (m *Manager) ListParts(id string) ([]meta.PartMeta, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, err := m.getLocked(id)
	if err != nil {
		return nil, err
	}
	out := make([]meta.PartMeta, 0, len(s.parts))
	for _, p := range s.parts {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out, nil
}

// writePartFile 写入分片文件：Sync 后 Close，再对外可见。
func writePartFile(ctx context.Context, path string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return errs.WrapCanceled(err)
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}

	// 先 Sync 刷盘，再 Close，最后 Rename 对外可见。
	// 顺序不能反：Close 后 fd 失效，再 Sync 无法落盘，
	// Rename 出去的就是未刷盘数据，对外会读到空分片或旧内容。
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := ctx.Err(); err != nil {
		_ = os.Remove(tmp)
		return errs.WrapCanceled(err)
	}
	return os.Rename(tmp, path)
}

// PartPath 导出分片路径（测试/排障）。
func (m *Manager) PartPath(id string, num int) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, err := m.getLocked(id)
	if err != nil {
		return "", err
	}
	return filepath.Join(s.dir, partName(num)), nil
}
