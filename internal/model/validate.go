package model

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalid  = errors.New("invalid input")
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("version conflict")
	ErrNotReady = errors.New("draft is not validated")
)

func (n Node) Validate() error {
	if strings.TrimSpace(n.ID) == "" || strings.TrimSpace(n.Name) == "" {
		return fmt.Errorf("%w: node id/name", ErrInvalid)
	}
	return nil
}
func (p Port) Validate() error {
	if p.ID == "" || p.NodeID == "" || p.RateBits <= 0 {
		return fmt.Errorf("%w: port fields", ErrInvalid)
	}
	return nil
}
func (l Link) Validate() error {
	if l.ID == "" || l.FromPort == "" || l.ToPort == "" || l.FromPort == l.ToPort || l.PropagationNS < 0 || l.PeriodNS <= 0 {
		return fmt.Errorf("%w: link fields", ErrInvalid)
	}
	return nil
}
func (s Stream) Validate() error {
	if s.ID == "" || s.PeriodNS <= 0 || s.FrameBits <= 0 || s.ReleaseNS < 0 || s.DeadlineNS <= 0 || s.MaxJitterNS < 0 || len(s.PathPorts) < 2 {
		return fmt.Errorf("%w: stream fields", ErrInvalid)
	}
	return nil
}
func (d ScheduleDraft) CanCommit() bool { return d.Status == Validated }
