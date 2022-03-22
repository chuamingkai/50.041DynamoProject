package models

type Object struct {
	Key string `json:"key"`
	Value string `json:"value"`
	VC     map[string]uint64 `json:",omitempty"`
}
