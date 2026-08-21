package etag

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
)

// Of 计算内容 ETag（sha256 hex）。
func Of(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// OfReader 流式计算。
func OfReader(r io.Reader) (string, int64, error) {
	h := sha256.New()
	n, err := io.Copy(h, r)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

// Multipart 组合多分片 ETag。
func Multipart(parts []string) string {
	h := sha256.New()
	for _, p := range parts {
		_, _ = io.WriteString(h, p)
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil)) + fmt.Sprintf("-%d", len(parts))
}

// NewHasher 暴露底层 hasher 供流式写。
func NewHasher() hash.Hash { return sha256.New() }
