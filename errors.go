package objmeta

import "example.com/objmeta/internal/errs"

var (
	ErrNotFound         = errs.ErrNotFound
	ErrAlreadyExists    = errs.ErrAlreadyExists
	ErrPrecondition     = errs.ErrPrecondition
	ErrClosed           = errs.ErrClosed
	ErrInvalidKey       = errs.ErrInvalidKey
	ErrInvalidPart      = errs.ErrInvalidPart
	ErrIncompleteParts  = errs.ErrIncompleteParts
	ErrUploadNotFound   = errs.ErrUploadNotFound
	ErrTooLarge         = errs.ErrTooLarge
	ErrCanceled         = errs.ErrCanceled
	ErrPersist          = errs.ErrPersist
)
