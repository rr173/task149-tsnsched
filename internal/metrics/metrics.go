package metrics

import (
	"sync/atomic"
	"time"
)

type Counters struct {
	Requests  atomic.Uint64
	Failures  atomic.Uint64
	Drafts    atomic.Uint64
	Commits   atomic.Uint64
	StartedAt time.Time
}

func New() *Counters { return &Counters{StartedAt: time.Now().UTC()} }

func (c *Counters) Request(ok bool) {
	c.Requests.Add(1)
	if !ok {
		c.Failures.Add(1)
	}
}

func (c *Counters) DraftCreated() { c.Drafts.Add(1) }
func (c *Counters) Committed()    { c.Commits.Add(1) }

func (c *Counters) Snapshot() map[string]any {
	return map[string]any{"requests": c.Requests.Load(), "failures": c.Failures.Load(), "drafts": c.Drafts.Load(), "commits": c.Commits.Load(), "started_at": c.StartedAt}
}
