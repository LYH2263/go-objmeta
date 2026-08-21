package meta

import "time"

// Clock 可注入时钟。
type Clock interface {
	Now() time.Time
}
