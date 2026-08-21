package meta

import (
	"time"

	"example.com/objmeta/internal/clone"
)

// ObjectMeta 持久化对象元数据。
type ObjectMeta struct {
	Key          string
	Size         int64
	ETag         string
	ContentType  string
	UserMeta     map[string]string
	LastModified time.Time
	Version      uint64
}

// UploadMeta 分片会话元数据。
type UploadMeta struct {
	ID          string
	Key         string
	ContentType string
	UserMeta    map[string]string
	InitiatedAt time.Time
}

// PartMeta 单个分片。
type PartMeta struct {
	Number int
	Size   int64
	ETag   string
}

// CloneMeta 深拷贝 ObjectMeta；UserMeta map 与原对象隔离。
func CloneMeta(m ObjectMeta) ObjectMeta {
	m.UserMeta = clone.StringMap(m.UserMeta)
	return m
}
