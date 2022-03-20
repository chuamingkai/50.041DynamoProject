package models

type Object struct {
	// IC     string            `json:"ic"`
	// GeoLoc string            `json:"geoloc"`
	Key string `json:"key"`
	Value string `json:"value"`
	VC     map[string]uint64 `json:",omitempty"`
}
