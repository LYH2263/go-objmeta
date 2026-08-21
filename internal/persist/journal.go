package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"example.com/objmeta/internal/errs"
)

// Journal 追加操作日志（审计/排障）。
type Journal struct {
	mu   sync.Mutex
	path string
	f    *os.File
}

type Entry struct {
	Time time.Time `json:"time"`
	Op   string    `json:"op"`
	Key  string    `json:"key"`
	Detail string  `json:"detail,omitempty"`
}

func OpenJournal(root string) (*Journal, error) {
	path := filepath.Join(root, "meta", "journal.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &Journal{path: path, f: f}, nil
}

func (j *Journal) Append(op, key, detail string) error {
	if j == nil {
		return nil
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	e := Entry{Time: time.Now().UTC(), Op: op, Key: key, Detail: detail}
	b, err := json.Marshal(e)
	if err != nil {
		return errs.WrapPersist("journal", err)
	}
	if _, err := j.f.Write(append(b, '\n')); err != nil {
		return errs.WrapPersist("journal", err)
	}
	return nil
}

func (j *Journal) Close() error {
	if j == nil || j.f == nil {
		return nil
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	err := j.f.Close()
	j.f = nil
	return err
}
