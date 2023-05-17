package loadbalance

import (
	"sync"

	"github.com/alphadose/haxmap"
)

type WeightedRoundRobin[T Hashable, I Instance[T]] struct {
	curIndex     int
	mutex        sync.Mutex
	instancesMap *haxmap.Map[T, *WeightNode[T, I]]
	instances    []*WeightNode[T, I]
}

type WeightNode[T Hashable, I Instance[T]] struct {
	instance        I
	Weight          int //初始化时对节点约定的权重
	currentWeight   int //节点临时权重，每轮都会变化
	effectiveWeight int //有效权重, 默认与weight相同 , totalWeight = sum(effectiveWeight)  //出现故障就-1
}

func NewWeightRoundRobin[T Hashable, I Instance[T]]() *WeightedRoundRobin[T, I] {
	return &WeightedRoundRobin[T, I]{
		instancesMap: haxmap.New[T, *WeightNode[T, I]](8),
		instances:    make([]*WeightNode[T, I], 0, 8),
	}
}

//1, currentWeight = currentWeight + effectiveWeight
//2, 选中最大的currentWeight节点为选中节点
//3, currentWeight = currentWeight - totalWeight

func (wrr *WeightedRoundRobin[T, I]) Add(instances ...I) int {
	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()
	count := 0
	for _, instance := range instances {
		id := instance.InstanceID()
		if _, ok := wrr.instancesMap.Get(id); !ok {
			node := &WeightNode[T, I]{
				instance: instance,
				Weight:   instance.InstanceWeight(),
			}
			node.effectiveWeight = node.Weight
			wrr.instancesMap.Set(id, node)
			wrr.instances = append(wrr.instances, node)
			count++
		}
	}

	return count
}

func (wrr *WeightedRoundRobin[T, I]) Get(key T) (I, bool) {
	//TODO implement me
	panic("implement me")
}

func (wrr *WeightedRoundRobin[T, I]) Select() (ins I) {
	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()

	var best *WeightNode[T, I]
	total := 0
	for i := 0; i < len(wrr.instances); i++ {
		w := wrr.instances[i]
		//1 计算所有有效权重
		total += w.effectiveWeight
		//2 修改当前节点临时权重
		w.currentWeight += w.effectiveWeight
		//3 有效权重默认与权重相同，通讯异常时-1, 通讯成功+1，直到恢复到weight大小
		if w.effectiveWeight < w.Weight {
			w.effectiveWeight++
		}
		//4 选中最大临时权重节点
		if best == nil || w.currentWeight > best.currentWeight {
			best = w
		}
	}
	if best == nil {
		return
	}
	//5 变更临时权重为 临时权重-有效权重之和
	best.currentWeight -= total
	return best.instance
}

func (wrr *WeightedRoundRobin[T, I]) Del(instances ...I) int {
	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()

	count := 0
	for _, instance := range instances {
		if _, ok := wrr.instancesMap.Get(instance.InstanceID()); ok {
			id := instance.InstanceID()
			for i := 0; i < len(wrr.instances); i++ {
				if wrr.instances[i].instance.InstanceID() == id {
					wrr.instancesMap.Del(id)
					wrr.instances = append(wrr.instances[:i], wrr.instances[i+1:]...)
					break
				}
			}
			count++
		}
	}

	return count
}

func (wrr *WeightedRoundRobin[T, I]) ForEach(callback func(T, I) bool) {
	haxMapForEach(wrr.instancesMap, func(key T, node *WeightNode[T, I]) bool {
		return callback(key, node.instance)
	})
}

func (wrr *WeightedRoundRobin[T, I]) Size() int {
	return int(wrr.instancesMap.Len())
}
