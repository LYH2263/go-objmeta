package store

import (
	"context"
	"io"

	"example.com/objmeta/internal/errs"
)

// CopyCancel 在拷贝过程中轮询 ctx。
func CopyCancel(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	buf := make([]byte, 32*1024)
	var written int64
	for {
		if err := ctx.Err(); err != nil {
			return written, errs.WrapCanceled(err)
		}
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[:nr])
			written += int64(nw)
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if er == io.EOF {
			return written, nil
		}
		if er != nil {
			return written, er
		}
	}
}
