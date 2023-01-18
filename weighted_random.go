package loadbalance

import (
	"sort"
	"sync"
	"time"

	"github.com/alphadose/haxmap"
)

type InstanceList[T Hashable, I Instance[T]] []I

func (ig InstanceList[T, I]) Len() int {
	return len(ig)
}

func (ig InstanceList[T, I]) Less(i, j int) bool {
	return ig[i].InstanceWeight() > ig[j].InstanceWeight()
}

func (ig InstanceList[T, I]) Swap(i, j int) {
	ig[i], ig[j] = ig[j], ig[i]
}

type WeightedRandom[T Hashable, I Instance[T]] struct {
	mutex        sync.Mutex
	instancesMap *haxmap.Map[T, I]
	instances    InstanceList[T, I]
	weightSum    int64
	random       XorShift64
	//random *rand.Rand
}

func NewWeightedRandom[T Hashable, I Instance[T]]() *WeightedRandom[T, I] {
	return &WeightedRandom[T, I]{
		instancesMap: haxmap.New[T, I](8),
		instances:    make(InstanceList[T, I], 0, 8),
		//random:       rand.New(rand.NewSource(time.Now().UnixNano())),
		random: XorShift64{
			state: uint64(time.Now().UnixNano()),
		},
	}
}

func (wr *WeightedRandom[T, I]) Add(instances ...I) int {
	wr.mutex.Lock()
	defer wr.mutex.Unlock()
	count := 0
	for _, instance := range instances {
		id := instance.InstanceID()
		if _, ok := wr.instancesMap.Get(id); !ok {
			wr.instancesMap.Set(id, instance)
			wr.instances = append(wr.instances, instance)
			wr.weightSum += int64(instance.InstanceWeight())
			count++
		}
	}
	if count > 0 {
		sort.Sort(wr.instances)
	}
	return count
}

func (wr *WeightedRandom[T, I]) Select() (ins I) {
	wr.mutex.Lock()
	defer wr.mutex.Unlock()
	rdm := wr.random.Int63()%wr.weightSum + 1
	//fmt.Println(rdm)
	if rdm < 0 {
		rdm = -rdm
	}
	for i := 0; i < len(wr.instances); i++ {
		rdm -= int64(wr.instances[i].InstanceWeight())

		if rdm <= 0 {
			return wr.instances[i]
		}
	}
	return
}

func (wr *WeightedRandom[T, I]) Get(key T) (I, bool) {
	return haxMapGetVal(wr.instancesMap, key)
}

func (wr *WeightedRandom[T, I]) Del(instances ...I) int {
	wr.mutex.Lock()
	defer wr.mutex.Unlock()
	count := 0

	for _, instance := range instances {
		if _, ok := wr.instancesMap.Get(instance.InstanceID()); ok {
			id := instance.InstanceID()
			for i := 0; i < len(wr.instances); i++ {
				if wr.instances[i].InstanceID() == id {
					wr.instancesMap.Del(instance.InstanceID())
					wr.instances = append(wr.instances[:i], wr.instances[i+1:]...)
					wr.weightSum -= int64(instance.InstanceWeight())
					break
				}
			}
			count++
		}
	}

	if count > 0 {
		sort.Sort(wr.instances)
	}
	return count
}

func (wr *WeightedRandom[T, I]) ForEach(callback func(T, I) bool) {
	haxMapForEach(wr.instancesMap, callback)
}

func (wr *WeightedRandom[T, I]) Size() int {
	return int(wr.instancesMap.Len())
}
