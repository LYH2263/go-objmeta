package multipart

import (
	"context"
	"os"

	"example.com/objmeta/internal/errs"
)

// Abort 先持久化删除意图再清目录；失败时会话仍可查。
func (m *Manager) Abort(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return errs.WrapCanceled(err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	s, err := m.getLocked(id)
	if err != nil {
		return err
	}
	// 先尝试删目录；成功后再从 map 移除。失败则保留会话，避免半清。
	if err := os.RemoveAll(s.dir); err != nil {
		return errs.WrapPersist("abort-remove", err)
	}
	delete(m.uploads, id)
	return nil
}
