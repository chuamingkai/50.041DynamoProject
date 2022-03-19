package models

type Object struct {
	IC     string            `json:"ic"`
	GeoLoc string            `json:"geoloc"`
	VC     map[string]uint64 `json:",omitempty"`
}
