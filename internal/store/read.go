package store

import (
	"context"
	"io"

	"example.com/objmeta/internal/errs"
)

// ReadAllLimited 读取 reader，尊重 ctx 与大小上限。
func ReadAllLimited(ctx context.Context, r io.Reader, max int64) ([]byte, error) {
	if r == nil {
		return nil, nil
	}
	var buf []byte
	tmp := make([]byte, 32*1024)
	var n int64
	for {
		if err := ctx.Err(); err != nil {
			return nil, errs.WrapCanceled(err)
		}
		nr, err := r.Read(tmp)
		if nr > 0 {
			n += int64(nr)
			if n > max {
				return nil, errs.ErrTooLarge
			}
			buf = append(buf, tmp[:nr]...)
		}
		if err == io.EOF {
			return buf, nil
		}
		if err != nil {
			return nil, err
		}
	}
}
