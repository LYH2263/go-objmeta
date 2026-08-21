package errs

import (
	"errors"
	"fmt"
)

// WrapIncomplete 保证 errors.Is(_, ErrIncompleteParts) 成立。
func WrapIncomplete(detail string) error {
	return fmt.Errorf("%w: %s", ErrIncompleteParts, detail)
}

// WrapPersist 包装持久化失败。
func WrapPersist(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %s: %v", ErrPersist, op, err)
}

// WrapCanceled 将 context 错误映射为可 Is 的取消哨兵。
func WrapCanceled(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", ErrCanceled, err)
}

// IsNotFound 便捷判断。
func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

// IsClosed 便捷判断。
func IsClosed(err error) bool { return errors.Is(err, ErrClosed) }
