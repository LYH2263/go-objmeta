package errs

import "errors"

var (
	ErrNotFound        = errors.New("objmeta: not found")
	ErrAlreadyExists   = errors.New("objmeta: already exists")
	ErrPrecondition    = errors.New("objmeta: precondition failed")
	ErrClosed          = errors.New("objmeta: store closed")
	ErrInvalidKey      = errors.New("objmeta: invalid key")
	ErrInvalidPart     = errors.New("objmeta: invalid part")
	ErrIncompleteParts = errors.New("objmeta: incomplete parts")
	ErrUploadNotFound  = errors.New("objmeta: upload not found")
	ErrTooLarge        = errors.New("objmeta: object too large")
	ErrCanceled        = errors.New("objmeta: canceled")
	ErrPersist         = errors.New("objmeta: persist failed")
)
