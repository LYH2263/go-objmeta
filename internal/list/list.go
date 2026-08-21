package list

import (
	"sort"
	"strings"

	"example.com/objmeta/internal/meta"
)

// Options 列举参数。
type Options struct {
	Prefix    string
	Delimiter string
	Limit     int
	Cursor    string
}

// Result 内部列举结果。
type Result struct {
	Entries        []meta.ObjectMeta
	CommonPrefixes []string
	NextCursor     string
	Truncated      bool
}

// Engine 复用 CommonPrefixes 底层数组；对外返回前必须深拷贝。
type Engine struct {
	prefs []string
}

// Filter 按前缀/分隔符过滤快照；CommonPrefixes 指向 Engine 内部缓冲。
func (e *Engine) Filter(all []meta.ObjectMeta, opt Options) Result {
	limit := opt.Limit
	if limit <= 0 {
		limit = 1000
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Key < all[j].Key })
	var entries []meta.ObjectMeta
	prefSet := map[string]struct{}{}
	e.prefs = e.prefs[:0]
	started := opt.Cursor == ""
	for _, m := range all {
		if opt.Prefix != "" && !strings.HasPrefix(m.Key, opt.Prefix) {
			continue
		}
		if !started {
			if m.Key <= opt.Cursor {
				continue
			}
			started = true
		}
		rest := m.Key[len(opt.Prefix):]
		if opt.Delimiter != "" {
			if i := strings.Index(rest, opt.Delimiter); i >= 0 {
				cp := opt.Prefix + rest[:i+len(opt.Delimiter)]
				if _, ok := prefSet[cp]; !ok {
					prefSet[cp] = struct{}{}
					e.prefs = append(e.prefs, cp)
				}
				continue
			}
		}
		entries = append(entries, meta.CloneMeta(m))
		if len(entries)+len(e.prefs) >= limit {
			return Result{
				Entries:        entries,
				CommonPrefixes: cloneStrings(e.prefs),
				NextCursor:     m.Key,
				Truncated:      true,
			}
		}
	}
	return Result{Entries: entries, CommonPrefixes: cloneStrings(e.prefs)}
}

// cloneStrings 深拷贝 e.prefs；防止调用方改写污染 Engine 复用缓冲。
func cloneStrings(src []string) []string {
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}
