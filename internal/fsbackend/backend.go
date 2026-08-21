package fsbackend

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"example.com/objmeta/internal/errs"
)

// Backend 文件数据面。
type Backend struct {
	root string
}

func Open(root string) (*Backend, error) {
	obj := filepath.Join(root, "objects")
	if err := os.MkdirAll(obj, 0o755); err != nil {
		return nil, err
	}
	return &Backend{root: root}, nil
}

func (b *Backend) Close() error { return nil }

func (b *Backend) objectPath(key string) string {
	return SafeJoin(filepath.Join(b.root, "objects"), key)
}

// WriteObject 原子写对象字节。
func (b *Backend) WriteObject(ctx context.Context, key string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return errs.WrapCanceled(err)
	}
	path := b.objectPath(key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	// 先写盘并 Sync，再 Close，最后 Rename；顺序错误会导致半写可见。
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
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

// OpenObject 打开只读流。
func (b *Backend) OpenObject(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, errs.WrapCanceled(err)
	}
	f, err := os.Open(b.objectPath(key))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	return f, nil
}

func (b *Backend) RemoveObject(key string) error {
	err := os.Remove(b.objectPath(key))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
