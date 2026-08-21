package validate

import "example.com/objmeta/internal/errs"

// PartNumber 校验分片号。
func PartNumber(n, max int) error {
	if n < 1 || n > max {
		return errs.ErrInvalidPart
	}
	return nil
}

// PartList 校验 complete 分片列表非空且无重复。
func PartList(parts []int) error {
	if len(parts) == 0 {
		return errs.ErrIncompleteParts
	}
	seen := make(map[int]struct{}, len(parts))
	for _, p := range parts {
		if p < 1 {
			return errs.ErrInvalidPart
		}
		if _, ok := seen[p]; ok {
			return errs.ErrInvalidPart
		}
		seen[p] = struct{}{}
	}
	return nil
}
