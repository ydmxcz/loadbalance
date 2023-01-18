# loadbalance

A Load Balance lib.

`WeightedDoubleQueue` using two queues implements a concurrency safe , fast , dynamic weighted load-balance  supported select, add and delete instance operation and these time complexity of operations all are O(1).The performance of this implementation is unrelated to the number of instance.

At first,define a struct named `instanceWrapper` which is includes two filed that one named `instance`(a type implements generic interface `Instance[T]`) and the other named `weight` (int).

## WeightDoubleQueue

It has exactly the same effect as normal weighted load balancing, but it performs better and is independent of the number of instances

- struct define

```go

type WeightedDoubleQueue[T Hashable, I Instance[T]] struct {
	// hashmap is a concurrency hash map
	hashmap *haxmap.Map[T, *instanceWrapper[T, I]]
	// mqueue is short for `main-queue`
	mqueue *queue[*instanceWrapper[T, I]]
	// squeue is short for `second-queue`
	squeue *queue[*instanceWrapper[T, I]]
	mutex  sync.Mutex
}
```

At first,define a struct named `instanceWrapper` which is includes two filed that one named `instance`(a type implements generic interface `Instance[T]`) and the other named `weight` (int).

```go
type instanceWrapper[T Hashable, I Instance[T]] struct {
	instance I
	weight   int
}
```

Then the step of select algorithm as follows:

1. pop node every time from main queue ,if the weight of current `instanceWrapper` is `minInt64` indicate that the instance was flagged delete,re-pop a Wrapper and give the current `instanceWrapper` to gc.
2. While the main queue is empty,swap the main queue and second queue.
3. After popping `instanceWrapper` that it didn't flag deleted,then the weight of current `instanceWrapper` subtract one,
4. if the weight is 0 ,reset the weight by the method of instance named `InstanceWeight()` , and push to the second queue,otherwise push to the main queue. At last,get and return the instance from the current  `instanceWrapper`.

## Some Load-Balance Implementations

two main interfaces

- Selector[T,I]
- SelectorBy[T,I]

as follows:

```go
type base[T Hashable, I Instance[T]] interface {
	Add(instances ...I) int
	Del(instances ...I) int
	Get(T) (I, bool)
	Size() int
	ForEach(func(T, I) bool)
}

type SelectorBy[T Hashable, I Instance[T]] interface {
	base[T, I]
	SelectBy(string) I
}

type Selector[T Hashable, I Instance[T]] interface {
	base[T, I]
	Select() I
}
```

some Implementations about `Selector[T,I]`

- WeightDoubleQueue

  > loadbalance.NewWeightDoubleQueue[T,Instance\[T\]\]()
  >
- Random Load Balance

  > loadbalance.NewRandom\[T,Instance\[T\]\]()
  >
- RoundRobin

  > loadbalance.NewRoundRobin[T,Instance\[T\]\]()
  >
- WeightedRandom

  > loadbalance.NewWeightedRandom[T,Instance\[T\]\]()
  >
- WeightedRoundRobin

  > loadbalance.NewWeightedRoundRobin[T,Instance\[T\]\]()
  >

some Implementations about `SelectorBy[T,I]`

- ConsistenceHash

  > loadbalance.NewConsistenceHash[T,Instance\[T\]\]()
  >
- SourceAddressHash

  > loadbalance.NewSourceAddressHash[T,Instance\[T\]\]()
  >

## Example

```go
package main

import (
	"fmt"
	"github.com/ydmxcz/loadbalance"
)

// define my instance struct and implement `Instance[string]`
type myService struct {
	Address         string
	Memory          int
}

func (ms *myService) InstanceID() string {
	return ms.Address
}

func (ms *myService) InstanceWeight() int {
	return ms.Memory
}

func main() {
	// define some service
	ins := []*myService{{
		Address: "192.168.0.90:9000",
		Memory:  5, // here memory as weight
	}, {
		Address: "192.168.0.90:9001",
		Memory:  3, // here memory as weight
	}, {
		Address: "192.168.0.90:9002",
		Memory:  2, // here memory as weight
	}}

	var selector loadbalance.Selector[string, *myService]

	// all methods is concurrncy safe and fast
	selector = loadbalance.NewWeightedDoubleQueue[string, *myService]()

	// add some instances 
	selector.Add(ins...)

	// select a instance
	_ = selector.Select()

	// delete a instances
	selector.Del(ins[2])

	// get a instance by its ID,the first result is instance,
	// second result is the instance wether exist
	_, exist := selector.Get("192.168.0.90:9002")

	size := selector.Size()

	// check all instance by for-each
	selector.ForEach(func(s string, ms *myService) bool {
		fmt.Println(ms)
		return true
	})
}


```

## Selector Parallel Benchmark Result

`-3`  indicate that there are 3 instances in `Selector`

`-16384 ` indicate that there are 16384 instances in `Selector`

- WeightDoubleQueue Benchmark Result

```
goos: windows
goarch: amd64
pkg: github.com/ydmxcz/loadbalance
cpu: Intel(R) Core(TM) i5-1035G1 CPU @ 1.00GHz
Benchmark_WeightedDoubleQueue
Benchmark_WeightedDoubleQueue/WeightedDoubleQueue_LoadBalance-3
Benchmark_WeightedDoubleQueue/WeightedDoubleQueue_LoadBalance-3-8
24424398                48.64 ns/op            0 B/op          0 allocs/op
Benchmark_WeightedDoubleQueue/WeightedDoubleQueue_LoadBalance-16384
Benchmark_WeightedDoubleQueue/WeightedDoubleQueue_LoadBalance-16384-8
22223332                52.59 ns/op            0 B/op          0 allocs/op
PASS
ok      github.com/ydmxcz/loadbalance   2.540s
```


- Other Load-Balance Implementations benchmark results

```plaintext
goos: windows
goarch: amd64
pkg: github.com/ydmxcz/loadbalance
cpu: Intel(R) Core(TM) i5-1035G1 CPU @ 1.00GHz
Benchmark_GetterLoadBalance_Get_Parallel
Benchmark_GetterLoadBalance_Get_Parallel/Random_LoadBalance-3
Benchmark_GetterLoadBalance_Get_Parallel/Random_LoadBalance-3-8
19799725                61.04 ns/op            0 B/op          0 allocs/op
Benchmark_GetterLoadBalance_Get_Parallel/Random_LoadBalance-16384
Benchmark_GetterLoadBalance_Get_Parallel/Random_LoadBalance-16384-8
20160124                59.80 ns/op            0 B/op          0 allocs/op
Benchmark_GetterLoadBalance_Get_Parallel/RoundRobin_LoadBalance-3
Benchmark_GetterLoadBalance_Get_Parallel/RoundRobin_LoadBalance-3-8
19869423                60.91 ns/op            0 B/op          0 allocs/op
Benchmark_GetterLoadBalance_Get_Parallel/RoundRobin_LoadBalance-16384
Benchmark_GetterLoadBalance_Get_Parallel/RoundRobin_LoadBalance-16384-8
19674452                61.09 ns/op            0 B/op          0 allocs/op
Benchmark_GetterLoadBalance_Get_Parallel/WeightedDoubleQueue_LoadBalance-3
Benchmark_GetterLoadBalance_Get_Parallel/WeightedDoubleQueue_LoadBalance-3-8
24103790                47.94 ns/op            0 B/op          0 allocs/op
Benchmark_GetterLoadBalance_Get_Parallel/WeightedDoubleQueue_LoadBalance-16384
Benchmark_GetterLoadBalance_Get_Parallel/WeightedDoubleQueue_LoadBalance-16384-8
22338339                52.23 ns/op            0 B/op          0 allocs/op
Benchmark_GetterLoadBalance_Get_Parallel/WeightRoundRobin_LoadBalance-3
Benchmark_GetterLoadBalance_Get_Parallel/WeightRoundRobin_LoadBalance-3-8
22139607                51.32 ns/op            0 B/op          0 allocs/op
Benchmark_GetterLoadBalance_Get_Parallel/WeightRoundRobin_LoadBalance-16384
Benchmark_GetterLoadBalance_Get_Parallel/WeightRoundRobin_LoadBalance-16384-8
   35388             33758 ns/op               0 B/op          0 allocs/op
Benchmark_GetterLoadBalance_Get_Parallel/WeightedRandom_LoadBalance-3
Benchmark_GetterLoadBalance_Get_Parallel/WeightedRandom_LoadBalance-3-8
19595419                60.90 ns/op            0 B/op          0 allocs/op
Benchmark_GetterLoadBalance_Get_Parallel/WeightedRandom_LoadBalance-16384
Benchmark_GetterLoadBalance_Get_Parallel/WeightedRandom_LoadBalance-16384-8
  162162              6963 ns/op               0 B/op          0 allocs/op
PASS
ok      github.com/ydmxcz/loadbalance   13.955s
```
