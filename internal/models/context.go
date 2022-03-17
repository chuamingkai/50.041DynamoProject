package models

type Context struct {
	//vectorClock      vclock.VClock
	isConflicting    bool   `bson:"isConflicting,omitempty"`
	writeCoordinator string `bson:"writeCoordinator,omitempty"`
}