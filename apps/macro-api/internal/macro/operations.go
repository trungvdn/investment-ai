package macro

import "time"

type SourceState struct {
	Provider    string     `json:"provider"`
	NextDue     time.Time  `json:"nextDue"`
	LeaseUntil  *time.Time `json:"leaseUntil"`
	LastChecked *time.Time `json:"lastChecked"`
	LastSuccess *time.Time `json:"lastSuccess"`
	Failures    int        `json:"consecutiveFailures"`
	Kind        string     `json:"errorKind"`
	Message     string     `json:"message"`
}
