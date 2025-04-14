package status

// Status represents the status of a transaction
type Status int

const (
	StatusPending Status = iota
	StatusConfirmed
	StatusFailed
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusConfirmed:
		return "confirmed"
	case StatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}
