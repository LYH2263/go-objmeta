package objmeta

import "example.com/objmeta/internal/meta"

// DefaultOptions 返回推荐默认值。
func DefaultOptions(root string) Options {
	return Options{
		Root:           root,
		MaxObjectBytes: 64 << 20,
		MaxParts:       10000,
	}
}

// WithClock 链式设置时钟。
func (o Options) WithClock(c meta.Clock) Options {
	o.Clock = c
	return o
}

// WithMaxBytes 链式设置对象大小上限。
func (o Options) WithMaxBytes(n int64) Options {
	o.MaxObjectBytes = n
	return o
}

// WithMaxParts 链式设置分片上限。
func (o Options) WithMaxParts(n int) Options {
	o.MaxParts = n
	return o
}
