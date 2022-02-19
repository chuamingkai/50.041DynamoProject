package main

import (
	"fmt"

	"github.com/DistributedClocks/GoVector/govec/vclock"
)

type Context struct {
	vectorClock      vclock.VClock
	isConflicting    bool
	writeCoordinator string
}

type Object struct {
	object  string  //geographical coordinates
	context Context //context of object
}

/*get()
  operation locates the object replicas associated with the key in the
  storage system and returns a single object or a list of objects with
  conflicting versions along with a context.*/

func GetObjects(key string) *Object {
	//call function to retrieve the node to contact
	//call all conflicting versions
	//return all conflicting versions and context (obj, context) pairs?
	return &Object{}
}

/*put(key, context,object)
  operation determines where the replicas of the object should be placed
  based on the associated key, and writes the replicas to disk.
  The context encodes system metadata about the object that is
  opaque to the caller and includes information such as
  the version of the object. The context information is stored along
  with the object so that the system can verify the validity of the
  context object supplied in the put request.*/

/*func PutObject(key string, obj Object) {

  }*/

func main() {
	var vc1 vclock.VClock = vclock.New()
	vc1.Set("123", 65535)
	fmt.Print(vc1["123"])
}
