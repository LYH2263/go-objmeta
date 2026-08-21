package objmeta

import (
	"example.com/objmeta/internal/admin"
	"example.com/objmeta/internal/meta"
)

// AdminSnapshot 供管理页一次性拉取。
func (s *Store) AdminSnapshot() admin.Dashboard {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return admin.Dashboard{}
	}
	snap := s.index.Snapshot()
	rows := admin.BuildObjects(snap, 200)
	var ups []admin.UploadRow
	// multipart 会话摘要：通过 Count 暴露规模，详细列表由 ListParts 路径覆盖
	_ = meta.ObjectMeta{}
	return admin.BuildDashboard(s.met.Snapshot(), rows, ups)
}

// Root 返回数据根目录。
func (s *Store) Root() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.root
}

// IsClosed 是否已关闭。
func (s *Store) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}
