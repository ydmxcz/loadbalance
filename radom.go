package loadbalance

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/alphadose/haxmap"
)

const (
	rngMax  = 1 << 63
	rngMask = rngMax - 1
)

type XorShift64 struct {
	state uint64
}

func NewXorShift64(seed uint64) XorShift64 {
	return XorShift64{
		state: seed,
	}
}

func (xs *XorShift64) Seed(seed uint64) {
	atomic.StoreUint64(&xs.state, seed)
}

func (xs *XorShift64) Uint64() uint64 {
	x := atomic.LoadUint64(&xs.state)
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	atomic.StoreUint64(&xs.state, x)
	return x
}

func (xs *XorShift64) Int63() int64 {
	return int64(xs.Uint64() & rngMask)
}

func (xs *XorShift64) Int31() int32 {
	return int32(int64(xs.Uint64()&rngMask) >> 32)
}

func NewRandom[T Hashable, I Instance[T]]() *Random[T, I] {
	return &Random[T, I]{
		instances:    make([]I, 0, 8),
		instancesMap: haxmap.New[T, I](8),
		random: XorShift64{
			state: uint64(time.Now().UnixNano()),
		},
	}
}

// Random 随机负载均衡
type Random[T Hashable, I Instance[T]] struct {
	mutex        sync.Mutex
	instancesMap *haxmap.Map[T, I] //map[T]Instance[T]
	instances    []I
	random       XorShift64
}

func (rb *Random[T, I]) Get(key T) (I, bool) {
	return haxMapGetVal(rb.instancesMap, key)
}

func (rb *Random[T, I]) ForEach(callback func(T, I) bool) {
	haxMapForEach(rb.instancesMap, callback)
}

func (rb *Random[T, I]) Add(instances ...I) int {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	count := 0
	for _, instance := range instances {
		if _, ok := rb.instancesMap.Get(instance.InstanceID()); !ok {
			rb.instancesMap.Set(instance.InstanceID(), instance)
			rb.instances = append(rb.instances, instance)
			count++
		}
	}
	return count
}

func (rb *Random[T, I]) Select() (ins I) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	if len(rb.instances) == 0 {
		return
	}
	return rb.instances[int(rb.random.Uint64()%uint64(len(rb.instances)))]
}

func (rb *Random[T, I]) Size() int {
	return int(rb.instancesMap.Len())
}

func (rb *Random[T, I]) Del(instances ...I) int {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	count := 0
	for _, instance := range instances {
		id := instance.InstanceID()
		if _, ok := rb.instancesMap.Get(id); ok {
			for i := 0; i < len(rb.instances); i++ {
				if rb.instances[i].InstanceID() == id {
					rb.instancesMap.Del(id)
					rb.instances = append(rb.instances[:i], rb.instances[i+1:]...)
					break
				}
			}

			count++
		}
	}
	return count

}
