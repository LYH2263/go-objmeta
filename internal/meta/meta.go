package meta

import "time"

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

// CloneMeta 深拷贝 ObjectMeta。
func CloneMeta(m ObjectMeta) ObjectMeta {
	um := make(map[string]string, len(m.UserMeta))
	for k, v := range m.UserMeta {
		um[k] = v
	}
	m.UserMeta = um
	return m
}
