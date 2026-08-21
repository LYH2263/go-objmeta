package validate

import (
	"strings"
	"unicode/utf8"

	"example.com/objmeta/internal/errs"
)

const MaxKeyLen = 1024

// Key 校验对象键。
func Key(key string) error {
	if key == "" || len(key) > MaxKeyLen {
		return errs.ErrInvalidKey
	}
	if !utf8.ValidString(key) {
		return errs.ErrInvalidKey
	}
	if strings.IndexByte(key, 92) >= 0 || strings.IndexByte(key, 0) >= 0 {
		return errs.ErrInvalidKey
	}
	if strings.HasPrefix(key, "/") || strings.Contains(key, "//") {
		return errs.ErrInvalidKey
	}
	parts := strings.Split(key, "/")
	for _, p := range parts {
		if p == "." || p == ".." || p == "" {
			return errs.ErrInvalidKey
		}
	}
	return nil
}

// Prefix 允许空前缀。
func Prefix(p string) error {
	if p == "" {
		return nil
	}
	if strings.IndexByte(p, 92) >= 0 {
		return errs.ErrInvalidKey
	}
	return nil
}
