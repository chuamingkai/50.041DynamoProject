package models

import (
	"time"
)

type Object struct {
	Key            string            `json:"key"`
	Value          string            `json:"value"`
	VC             map[string]uint64 `json:"VC"`
	CreatedOn      time.Time         `json:",omitempty"`
	LastModifiedOn time.Time         `json:",omitempty"`
}

type HintedObject struct {
	Data       Object
	BucketName string
}
