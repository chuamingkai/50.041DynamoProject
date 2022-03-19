package models

// import "github.com/chuamingkai/50.041DynamoProject/pkg/vectorclock"

type Context struct {
	// VC 					vectorclock.VectorClock `bson:"vectorClock,omitempty"`
	IsConflicting    	bool   					`bson:"isConflicting,omitempty"`
	WriteCoordinator 	string 					`bson:"writeCoordinator,omitempty"`
}