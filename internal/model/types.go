package model

import "time"

type Status string

const (
	Draft      Status = "draft"
	Validated  Status = "validated"
	Committed  Status = "committed"
	Rejected   Status = "rejected"
	RolledBack Status = "rolled_back"
)

type ViolationKind string

const (
	Conflict     ViolationKind = "conflict"
	PathError    ViolationKind = "path_error"
	Deadline     ViolationKind = "deadline"
	Jitter       ViolationKind = "jitter"
	VersionError ViolationKind = "version"
)

type Direction string

const (
	Ingress       Direction = "ingress"
	Egress        Direction = "egress"
	Bidirectional Direction = "bidirectional"
)

type Node struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ClockDomain string    `json:"clock_domain"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
}
type Port struct {
	ID        string    `json:"id"`
	NodeID    string    `json:"node_id"`
	Name      string    `json:"name"`
	Direction Direction `json:"direction"`
	RateBits  int64     `json:"rate_bits"`
	CreatedAt time.Time `json:"created_at"`
}
type Link struct {
	ID            string    `json:"id"`
	FromPort      string    `json:"from_port"`
	ToPort        string    `json:"to_port"`
	PropagationNS int64     `json:"propagation_ns"`
	PeriodNS      int64     `json:"period_ns"`
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
}
type Stream struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PeriodNS    int64     `json:"period_ns"`
	FrameBits   int64     `json:"frame_bits"`
	ReleaseNS   int64     `json:"release_ns"`
	DeadlineNS  int64     `json:"deadline_ns"`
	MaxJitterNS int64     `json:"max_jitter_ns"`
	PathPorts   []string  `json:"path_ports"`
	CreatedAt   time.Time `json:"created_at"`
}
type ScheduleDraft struct {
	ID        string    `json:"id"`
	NetworkID string    `json:"network_id"`
	Version   int64     `json:"version"`
	Status    Status    `json:"status"`
	PeriodNS  int64     `json:"period_ns"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type Allocation struct {
	ID            string `json:"id"`
	DraftID       string `json:"draft_id"`
	StreamID      string `json:"stream_id"`
	LinkID        string `json:"link_id"`
	PortID        string `json:"port_id"`
	StartNS       int64  `json:"start_ns"`
	EndNS         int64  `json:"end_ns"`
	GuardBeforeNS int64  `json:"guard_before_ns"`
	GuardAfterNS  int64  `json:"guard_after_ns"`
	ArrivalNS     int64  `json:"arrival_ns"`
	DepartureNS   int64  `json:"departure_ns"`
}
type Violation struct {
	ID        string        `json:"id"`
	DraftID   string        `json:"draft_id"`
	Kind      ViolationKind `json:"kind"`
	StreamA   string        `json:"stream_a"`
	StreamB   string        `json:"stream_b"`
	PortID    string        `json:"port_id"`
	Detail    string        `json:"detail"`
	Resolved  bool          `json:"resolved"`
	CreatedAt time.Time     `json:"created_at"`
}
type ValidationResult struct {
	DraftID    string      `json:"draft_id"`
	Valid      bool        `json:"valid"`
	Violations []Violation `json:"violations"`
	CheckedAt  time.Time   `json:"checked_at"`
}
type ActiveVersion struct {
	NetworkID string    `json:"network_id"`
	DraftID   string    `json:"draft_id"`
	Version   int64     `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}
type ScheduleSummary struct {
	Draft       ScheduleDraft    `json:"draft"`
	Allocations []Allocation     `json:"allocations"`
	Validation  ValidationResult `json:"validation"`
}
type Window struct {
	FromNS      int64        `json:"from_ns"`
	ToNS        int64        `json:"to_ns"`
	Allocations []Allocation `json:"allocations"`
}
