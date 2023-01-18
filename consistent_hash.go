package loadbalance

import (
	"fmt"
	"sort"
	"sync"

	"github.com/alphadose/haxmap"
)

// 1 单调性（唯一） 2平衡性 (数据 目标元素均衡) 3分散性(散列)
type strHashFunc func(data string) uint64

type uint64Slice []uint64

func (s uint64Slice) Len() int {
	return len(s)
}

func (s uint64Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s uint64Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type ConsistentHash[T Hashable] struct {
	mux         sync.RWMutex
	hashfunc    strHashFunc
	replicas    int                              //复制因子
	keys        uint64Slice                      //已排序的节点hash切片
	keyMap      *haxmap.Map[uint64, Instance[T]] //节点哈希和key的map, 键是hash值，值是节点key
	instanceMap *haxmap.Map[T, Instance[T]]      //节点哈希和key的map, 键是hash值，值是节点key
}

func NewConsistentHash[T Hashable](replicas ...int) *ConsistentHash[T] {
	r := 3
	if len(replicas) != 0 {
		r = replicas[0]
	}
	m := &ConsistentHash[T]{
		replicas:    r,
		hashfunc:    GetHashFunc[string](),
		keyMap:      haxmap.New[uint64, Instance[T]](),
		instanceMap: haxmap.New[T, Instance[T]](),
	}
	return m
}

func (c *ConsistentHash[T]) Get(key T) (Instance[T], bool) {
	return c.instanceMap.Get(key)
}

func (c *ConsistentHash[T]) ForEach(callback func(T, Instance[T]) bool) {
	haxMapForEach(c.instanceMap, callback)
}

// Add 方法用来添加缓存节点，参数为节点key，比如使用IP
func (c *ConsistentHash[T]) Add(instances ...Instance[T]) int {
	c.mux.Lock()
	defer c.mux.Unlock()
	count := 0
	flag := true
	for _, instance := range instances {
		flag = true
		id := instance.InstanceID()
		for i := 0; i < c.replicas; i++ {

			if _, ok := c.instanceMap.GetOrSet(id, instance); !ok {
				hash := c.hashfunc(fmt.Sprintf("%v-%d", id, i))
				c.keys = append(c.keys, hash)
				c.keyMap.Set(hash, instance)
			} else {
				flag = false
				break
			}
		}
		if flag {
			count++
		}
	}
	sort.Sort(c.keys)
	return count
}

func (c *ConsistentHash[T]) delSlice(val uint64) {
	for i := 0; i < len(c.keys); i++ {
		if c.keys[i] == val {
			c.keys = append(c.keys[:i], c.keys[i+1:]...)
			break
		}
	}
}

func (c *ConsistentHash[T]) Size() int {
	return int(c.keyMap.Len())
}

func (c *ConsistentHash[T]) Del(instances ...Instance[T]) int {
	c.mux.Lock()
	defer c.mux.Unlock()
	count := 0
	flag := true
	for _, instance := range instances {
		flag = true
		id := instance.InstanceID()
		for i := 0; i < c.replicas; i++ {
			if _, ok := c.instanceMap.Get(id); ok {
				c.instanceMap.Del(id)
				hash := c.hashfunc(fmt.Sprintf("%v-%d", id, i))
				c.keyMap.Del(hash)
				c.delSlice(hash)
			} else {
				flag = false
				break
			}
		}
		if flag {
			count++
		}
	}
	return count
}

// Select 方法根据给定的对象获取最靠近它的那个节点
func (c *ConsistentHash[T]) Select(key string) Instance[T] {
	if c.keyMap.Len() == 0 {
		return nil
	}
	hash := c.hashfunc(key)
	c.mux.RLock()
	defer c.mux.RUnlock()
	idx := sort.Search(len(c.keys), func(i int) bool { return c.keys[i] >= hash })
	k := c.keys[idx]

	if idx == len(c.keys) {
		idx = 0
	}
	ins, _ := c.keyMap.Get(k)
	return ins
}
