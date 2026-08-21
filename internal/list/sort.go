package list

import (
	"sort"

	"example.com/objmeta/internal/meta"
)

// SortByKey 就地排序。
func SortByKey(items []meta.ObjectMeta) {
	sort.Slice(items, func(i, j int) bool { return items[i].Key < items[j].Key })
}

// SortKeys 排序字符串键。
func SortKeys(keys []string) {
	sort.Strings(keys)
}
