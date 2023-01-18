package loadbalance

import (
	"sync"

	"github.com/alphadose/haxmap"
)

// 轮询负载均衡
type RoundRobin[T Hashable, I Instance[T]] struct {
	curIndex     int
	mutex        sync.Mutex
	instancesMap *haxmap.Map[T, I] //map[T]I
	instances    []I
}

func NewRoundRobin[T Hashable, I Instance[T]]() *RoundRobin[T, I] {
	return &RoundRobin[T, I]{
		instancesMap: haxmap.New[T, I](8),
		instances:    make([]I, 0, 8),
	}
}

func (rr *RoundRobin[T, I]) Add(instances ...I) int {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	count := 0
	for _, instance := range instances {
		if _, ok := rr.instancesMap.Get(instance.InstanceID()); !ok {
			rr.instancesMap.Set(instance.InstanceID(), instance)

			rr.instances = append(rr.instances, instance)
			count++
		}

	}
	return count
}

func haxMapGetVal[K Hashable, V any](m *haxmap.Map[K, V], key K) (V, bool) {
	return m.Get(key)
}

func haxMapForEach[K Hashable, V any](m *haxmap.Map[K, V], callback func(K, V) bool) {
	m.ForEach(callback)
}

func (rr *RoundRobin[T, I]) Get(key T) (I, bool) {
	return haxMapGetVal(rr.instancesMap, key)
}

func (rr *RoundRobin[T, I]) ForEach(callback func(T, I) bool) {
	haxMapForEach(rr.instancesMap, callback)
}

func (rr *RoundRobin[T, I]) Select() (ins I) {
	if rr.instancesMap.Len() == 0 {
		return ins
	}

	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	c := (rr.curIndex + 1) % len(rr.instances)
	rr.curIndex = c
	return rr.instances[c]
}

func (rr *RoundRobin[T, I]) Del(instances ...I) int {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()
	count := 0
	for _, instance := range instances {
		id := instance.InstanceID()
		if _, ok := rr.instancesMap.Get(id); ok {
			for i := 0; i < len(rr.instances); i++ {
				if rr.instances[i].InstanceID() == id {
					rr.instancesMap.Del(instance.InstanceID())
					rr.instances = append(rr.instances[:i], rr.instances[i+1:]...)
					break
				}
			}
			count++
		}
	}
	return count
}

func (rr *RoundRobin[T, I]) Size() int {
	return int(rr.instancesMap.Len())
}
