package models

type GetContext struct {
	HasConflicts bool `json:"hasConflicts"`
	WriteCoordinator int `json:"writeCoordinator"`
	VectorClock map[string]uint64 `json:"vectorClock,omitempty"`
}