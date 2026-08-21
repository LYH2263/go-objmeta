package admin

import (
	"sort"
	"time"

	"example.com/objmeta/internal/meta"
	"example.com/objmeta/internal/metrics"
)

// ObjectRow 管理页列表行。
type ObjectRow struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	ETag         string    `json:"etag"`
	ContentType  string    `json:"content_type"`
	LastModified time.Time `json:"last_modified"`
}

// UploadRow 分片会话行。
type UploadRow struct {
	UploadID    string    `json:"upload_id"`
	Key         string    `json:"key"`
	InitiatedAt time.Time `json:"initiated_at"`
	PartCount   int       `json:"part_count"`
}

// Dashboard 汇总。
type Dashboard struct {
	Stats    metrics.Snapshot `json:"stats"`
	Objects  []ObjectRow     `json:"objects"`
	Uploads  []UploadRow      `json:"uploads"`
	Generated time.Time       `json:"generated"`
}

// BuildObjects 从元数据构建表格。
func BuildObjects(items []meta.ObjectMeta, limit int) []ObjectRow {
	sort.Slice(items, func(i, j int) bool { return items[i].Key < items[j].Key })
	if limit <= 0 || limit > len(items) {
		limit = len(items)
	}
	out := make([]ObjectRow, 0, limit)
	for i := 0; i < limit; i++ {
		m := items[i]
		out = append(out, ObjectRow{
			Key: m.Key, Size: m.Size, ETag: m.ETag,
			ContentType: m.ContentType, LastModified: m.LastModified,
		})
	}
	return out
}

// BuildDashboard 组装管理页数据。
func BuildDashboard(stats metrics.Snapshot, objs []ObjectRow, ups []UploadRow) Dashboard {
	return Dashboard{
		Stats: stats, Objects: objs, Uploads: ups, Generated: time.Now().UTC(),
	}
}
