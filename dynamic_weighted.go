package loadbalance

import (
	"sync"

	"github.com/alphadose/haxmap"
)

const minInt64 = -1 << 63

type queueNode[T any] struct {
	val  T
	next *queueNode[T]
}

func newNode[T any](val T) *queueNode[T] {
	return &queueNode[T]{
		val:  val,
		next: nil,
	}
}

// queue is a simply linked queue just 16 byte
// the basic unit of dequeue and enqueue is the `queueNode`
type queue[T any] struct {
	head *queueNode[T]
	tail *queueNode[T]
}

func (q *queue[T]) push(node *queueNode[T]) {
	// assign nil to `node.next` are needed
	node.next = nil
	if q.head == nil {
		q.head = node
		q.tail = node
	} else {
		q.tail.next = node
		q.tail = node
	}
}

func (q *queue[T]) pop() *queueNode[T] {
	n := q.head
	if n == nil {
		return nil
	}
	q.head = n.next
	if q.head == nil {
		q.tail = nil
	}
	return n
}

// Empty returns the queue wether empty
func (q *queue[T]) Empty() bool {
	return q.head == nil
}

// `instanceWrapper` which is includes two fileds
// that one named `instance`(a type implements generic interface `Instance[T]`)
// and the other named `weight` (int).
type instanceWrapper[T Hashable, I Instance[T]] struct {
	instance I
	weight   int
}

// DynamicWeighted uses two queue implements
// a concurrency safe , fast , dynamic weighted load-balance  supported select,
// add and delete instance operation and these time complexity of operations all are O(1).
// The performance of this implementation is unrelated to the number of instance.
//
// At first,define a struct named `instanceWrapper` which is includes two filed that
// one named `instance`(a type implements generic interface `Instance[T]`)
// and the other named `weight` (int).
//
// Then the step of select algorithm as follows:
//  1. pop node every time from this queue
//     if the weight of current `instanceWrapper` is `minInt64` ,
//     indicate that the instance was flagged delete,
//     re-pop a Wrapper and give the current `instanceWrapper` to gc.
//  2. While the main queue is empty,swap the main queue and second queue.
//  3. After popping `instanceWrapper` that it didn't flag deleted,
//     then the weight of current `instanceWrapper` subtract one,
//  4. if the weight is 0 ,
//     reset the weight by the method of instance named `InstanceWeight()` ,
//     and push to the second queue,
//     otherwise push to the main queue.
//     At last,get and return the instance from the current  `instanceWrapper`.
//
// some notes :
//  1. type I better is a type that can be compared with `nil`,
//     such as a pointer or an interface.
//  1. this processing needs the protection of the mutex lock,
//     After doing many implements includes using
//     lock-free queue and atomic operations and
//     doing more times test and benchmark,the performance of current way is the best.
//  2. Maybe my implementation used lock-free queue has some problem,
//     the result of implementation that used lock-free queue
//     and atomic operation of benchmark test always about 180 ns/op.
type DynamicWeighted[T Hashable, I Instance[T]] struct {
	// hashmap is a concurrency hash map
	hashmap *haxmap.Map[T, *instanceWrapper[T, I]]
	// mqueue is short for `main-queue`
	// note that :
	// Although the size of queue just 16 bytes and seems like friendly for GC，
	// based the result of parallel benchmark,
	// `Select` method will slow about 5 nanosecond every operation in my computer
	// if the type of queue is value type. so here still is reference type(a pointer).
	mqueue *queue[*instanceWrapper[T, I]]
	// squeue is short for `second-queue`
	squeue *queue[*instanceWrapper[T, I]]
	mutex  sync.Mutex
}

func NewDynamicWeighted[T Hashable, I Instance[T]]() *DynamicWeighted[T, I] {
	return &DynamicWeighted[T, I]{
		hashmap: haxmap.New[T, *instanceWrapper[T, I]](),
		mqueue:  &queue[*instanceWrapper[T, I]]{},
		squeue:  &queue[*instanceWrapper[T, I]]{},
	}
}

// Add some instances and return the number of successful operation
func (sl *DynamicWeighted[T, I]) Add(instances ...I) int {
	count := 0
	// traverse all instances
	for _, instance := range instances {
		// get the weight and unique-id of instance
		instanceId := instance.InstanceID()
		instanceWeight := instance.InstanceWeight()
		// query from hashmap,do add operator if not exit
		_, ok := sl.hashmap.Get(instanceId)
		if ok {
			continue
		} else {
			iw := &instanceWrapper[T, I]{
				instance: instance,
				weight:   instanceWeight,
			}
			n := newNode(iw)

			sl.mutex.Lock()
			sl.mqueue.push(n)
			sl.mutex.Unlock()

			sl.hashmap.Set(instanceId, iw)

		}
		count++
	}
	return count

}

// Del some instances and return the number of successful operation
func (sl *DynamicWeighted[T, I]) Del(instances ...I) int {
	// sl.hashmap is concurrency safe,the mutex is for simply
	// the problem of safely updating in this method and safely
	// reading in the method named `Select` the 'weight' filed in the type of `instanceWarrper`
	// here if no the mutex, there are more atomic operations about the `weight` filed in the `Select` method.
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	count := 0
	for _, instance := range instances {
		instanceId := instance.InstanceID()
		iw, ok := sl.hashmap.Get(instanceId)
		if !ok {
			continue
		} else {
			sl.hashmap.Del(instanceId)
			// atomic.StoreInt64(&iw.weight, minInt64)
			iw.weight = minInt64
		}
		count++
	}
	return count
}

func (sl *DynamicWeighted[T, I]) Size() int {
	return int(sl.hashmap.Len())
}

// ForEach every instances. it is concurrency safe.
func (sl *DynamicWeighted[T, I]) ForEach(callback func(T, I) bool) {
	haxMapForEach(sl.hashmap, func(key T, node *instanceWrapper[T, I]) bool {
		return callback(key, node.instance)
	})
}

// Get the value corresponding to the key
func (sl *DynamicWeighted[T, I]) Get(key T) (ins I, ok bool) {
	if w, ok := haxMapGetVal(sl.hashmap, key); ok {
		return w.instance, true
	}
	return
}

// Select a instance
func (sl *DynamicWeighted[T, I]) Select() (ins I) {
	// 互斥锁加锁
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
repop:
	// 出队一个节点
	instNode := sl.mqueue.pop()
	// 如果此节点为空，则mqueue为空
	if instNode == nil {
		// 若squeue也同时为空，则说明负载均衡器没有实例，返回
		if sl.squeue.Empty() {
			return
		}
		// 交换 mqueue 和 squeue
		sl.mqueue, sl.squeue = sl.squeue, sl.mqueue
		// 使用goto开启下一次循环
		goto repop
	}
	// 从节点中取出实例
	inst := instNode.val
	// 该实例已被删除
	if inst.weight == minInt64 {
		// 重新出队一个节点
		goto repop
	}
	inst.weight--
	if inst.weight == 0 {
		// 重新获取权重
		inst.weight = inst.instance.InstanceWeight()
		sl.squeue.push(instNode)
	} else {
		sl.mqueue.push(instNode)
	}
	// 返回实例
	return inst.instance
}
