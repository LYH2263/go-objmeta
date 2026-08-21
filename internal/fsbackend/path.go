package fsbackend

import (
	"path/filepath"
	"strings"
)

// SafeJoin 防止路径穿越。
func SafeJoin(root, key string) string {
	clean := filepath.Clean("/" + strings.ReplaceAll(key, "\\", "/"))
	rel := strings.TrimPrefix(clean, "/")
	return filepath.Join(root, filepath.FromSlash(rel))
}

// RelKey 从绝对路径还原 key（尽力）。
func RelKey(root, abs string) (string, error) {
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}
