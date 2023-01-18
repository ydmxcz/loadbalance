package loadbalance

import (
	"sync"

	"github.com/alphadose/haxmap"
)

type SourceAddressHash[T Hashable] struct {
	instanceMap *haxmap.Map[T, Instance[T]]
	insList     []Instance[T]
	hashfunc    strHashFunc
	rwmutex     sync.RWMutex
}

func NewSourceAddressHash[T Hashable]() *SourceAddressHash[T] {
	return &SourceAddressHash[T]{
		instanceMap: haxmap.New[T, Instance[T]](8),
		insList:     make([]Instance[T], 0, 8),
		hashfunc:    GetHashFunc[string](),
	}
}

func (kh *SourceAddressHash[T]) delSlice(val T) {
	for i := 0; i < len(kh.insList); i++ {
		if kh.insList[i].InstanceID() == val {
			kh.insList = append(kh.insList[:i], kh.insList[i+1:]...)
			break
		}
	}
}

func (kh *SourceAddressHash[T]) ForEach(callback func(T, Instance[T]) bool) {
	haxMapForEach(kh.instanceMap, callback)
}

func (kh *SourceAddressHash[T]) Del(instances ...Instance[T]) int {
	count := 0
	for _, instance := range instances {
		id := instance.InstanceID()
		if _, ok := kh.instanceMap.Get(id); !ok {
			kh.rwmutex.Lock()

			kh.delSlice(id)
			kh.instanceMap.Del(id)

			kh.rwmutex.Unlock()
			count++
		}
	}
	return count
}

func (kh *SourceAddressHash[T]) Add(instances ...Instance[T]) int {

	count := 0
	for _, instance := range instances {
		if _, ok := kh.instanceMap.Get(instance.InstanceID()); !ok {
			kh.rwmutex.Lock()
			kh.instanceMap.Set(instance.InstanceID(), instance)
			kh.insList = append(kh.insList, instance)
			kh.rwmutex.Unlock()
			count++
		}
	}
	return count

}

func (kh *SourceAddressHash[T]) Get(key T) (Instance[T], bool) {
	return haxMapGetVal(kh.instanceMap, key)
}

func (kh *SourceAddressHash[T]) SelectBy(key string) Instance[T] {
	idx := kh.hashfunc(key) % uint64(kh.instanceMap.Len())
	kh.rwmutex.RLock()
	defer kh.rwmutex.RUnlock()
	return kh.insList[idx]
}

func (kh *SourceAddressHash[T]) Size() int {
	return int(kh.instanceMap.Len())
}
