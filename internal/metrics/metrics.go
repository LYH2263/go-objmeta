package metrics

import "sync/atomic"

// Registry 原子计数器。
type Registry struct {
	puts, gets, deletes atomic.Uint64
	mpCreate, mpComplete, mpAbort, parts atomic.Uint64
}

func New() *Registry { return &Registry{} }

func (r *Registry) IncPut()               { r.puts.Add(1) }
func (r *Registry) IncGet()               { r.gets.Add(1) }
func (r *Registry) IncDelete()            { r.deletes.Add(1) }
func (r *Registry) IncMultipartCreate()   { r.mpCreate.Add(1) }
func (r *Registry) IncMultipartComplete() { r.mpComplete.Add(1) }
func (r *Registry) IncMultipartAbort()    { r.mpAbort.Add(1) }
func (r *Registry) IncPart()              { r.parts.Add(1) }

// Snapshot 只读快照。
type Snapshot struct {
	Puts              uint64 `json:"puts"`
	Gets              uint64 `json:"gets"`
	Deletes           uint64 `json:"deletes"`
	MultipartCreate   uint64 `json:"multipart_create"`
	MultipartComplete uint64 `json:"multipart_complete"`
	MultipartAbort    uint64 `json:"multipart_abort"`
	Parts             uint64 `json:"parts"`
}

func (r *Registry) Snapshot() Snapshot {
	return Snapshot{
		Puts:              r.puts.Load(),
		Gets:              r.gets.Load(),
		Deletes:           r.deletes.Load(),
		MultipartCreate:   r.mpCreate.Load(),
		MultipartComplete: r.mpComplete.Load(),
		MultipartAbort:    r.mpAbort.Load(),
		Parts:             r.parts.Load(),
	}
}
