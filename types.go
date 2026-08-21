package objmeta

import (
	"io"
	"time"

	"example.com/objmeta/internal/meta"
)

// ObjectInfo 对外可见的对象头信息（不含数据面）。
type ObjectInfo struct {
	Key          string
	Size         int64
	ETag         string
	ContentType  string
	UserMeta     map[string]string
	LastModified time.Time
	Version      uint64
}

// Object 含可读数据流；Close 释放底层句柄。
type Object struct {
	Info ObjectInfo
	Body io.ReadCloser
}

// PutOptions 控制覆盖与条件写。
type PutOptions struct {
	ContentType string
	UserMeta    map[string]string
	IfNoneMatch bool
	IfMatchETag string
}

// ListOptions 前缀列举。
type ListOptions struct {
	Prefix    string
	Delimiter string
	Limit     int
	Cursor    string
}

// ListResult 列举结果；Entries 为拷贝，调用方可安全修改。
type ListResult struct {
	Entries    []ObjectInfo
	CommonPrefs []string
	NextCursor string
	Truncated  bool
}

// MultipartUpload 分片会话句柄。
type MultipartUpload struct {
	UploadID    string
	Key         string
	InitiatedAt time.Time
}

// MultipartOptions 初始化分片上传。
type MultipartOptions struct {
	ContentType string
	UserMeta    map[string]string
}

// PartInfo 已上传分片摘要。
type PartInfo struct {
	PartNumber int
	Size       int64
	ETag       string
}

// Options 打开 Store。
type Options struct {
	Root           string
	MaxObjectBytes int64
	MaxParts       int
	Clock          meta.Clock
}

func toPublicInfo(m meta.ObjectMeta) ObjectInfo {
	um := make(map[string]string, len(m.UserMeta))
	for k, v := range m.UserMeta {
		um[k] = v
	}
	return ObjectInfo{
		Key:          m.Key,
		Size:         m.Size,
		ETag:         m.ETag,
		ContentType:  m.ContentType,
		UserMeta:     um,
		LastModified: m.LastModified,
		Version:      m.Version,
	}
}
