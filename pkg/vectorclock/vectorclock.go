package vectorclock

import (
	"github.com/DistributedClocks/GoVector/govec/vclock"
)

/*wrapper function to increase vector clock value of a particular index by 1*/
/*i.e. vc1={"A":1}, update("B",vc1)={"A":1,"B":1}*/
func UpdateRecv(index string, orig map[string]uint64) map[string]uint64 {
	updated := vclock.New().CopyFromMap(orig)
	old, check := updated.FindTicks(index)
	if !check {
		updated.Set(index, 1)
	} else {
		updated.Set(index, old+1)
	}
	return updated.GetMap()
}