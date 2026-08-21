package policy

import "example.com/objmeta/internal/errs"

// Guard 条件写策略。
type Guard struct{}

func New() *Guard { return &Guard{} }

// CheckPut 检查 If-None-Match / If-Match。
func (g *Guard) CheckPut(exists bool, existETag string, ifNoneMatch bool, ifMatch string) error {
	if ifNoneMatch && exists {
		return errs.ErrAlreadyExists
	}
	if ifMatch != "" {
		if !exists {
			return errs.ErrPrecondition
		}
		if existETag != ifMatch {
			return errs.ErrPrecondition
		}
	}
	return nil
}

// AllowOverwrite 默认允许覆盖。
func (g *Guard) AllowOverwrite() bool { return true }
