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

// Check if clock 2 is an ancestor of clock 1
func IsAncestorOf(clock1Map, clock2Map map[string]uint64) bool {
	clock1 := vclock.New().CopyFromMap(clock1Map)
	clock2 := vclock.New().CopyFromMap(clock2Map)

	return clock1.Compare(clock2, vclock.Ancestor)
}

// Check if clock 2 is a descendant of clock 1
func IsDescendantOf(clock1Map, clock2Map map[string]uint64) bool {
	clock1 := vclock.New().CopyFromMap(clock1Map)
	clock2 := vclock.New().CopyFromMap(clock2Map)

	return clock1.Compare(clock2, vclock.Descendant)
}

// Check if clock 2 is equal to clock 1
func IsEqualTo(clock1Map, clock2Map map[string]uint64) bool {
	clock1 := vclock.New().CopyFromMap(clock1Map)
	clock2 := vclock.New().CopyFromMap(clock2Map)

	return clock1.Compare(clock2, vclock.Equal)
}

// Check if clock 2 is concurrent with clock 1
func IsConcurrentWith(clock1Map, clock2Map map[string]uint64) bool {
	clock1 := vclock.New().CopyFromMap(clock1Map)
	clock2 := vclock.New().CopyFromMap(clock2Map)

	return clock1.Compare(clock2, vclock.Concurrent)
}

// Merge clock 2 into clock 1
func MergeClocks(clock1Map, clock2Map map[string]uint64) map[string]uint64 {
	clock1 := vclock.New().CopyFromMap(clock1Map)
	clock2 := vclock.New().CopyFromMap(clock2Map)

	clock1.Merge(clock2)

	return clock1.GetMap()
}