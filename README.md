# Loadbalance

A Load Balance lib for Go.

Implement * Faster Loadbalance Algorithm named DynamicWeighted *

`DynamicWeighted` using two queues implements a concurrency safe , fast , dynamic weighted load-balance  supported select, add and delete instance operation and these time complexity of operations all are O(1).The performance of this implementation is unrelated to the number of instance.


## DynamicWeight Loadbalance

It has exactly the same effect as normal weighted load balancing, but it performs better and is independent of the number of instances

## Here includes other loadbalance algorithm
- Consistent Hash
- Random
- RoundRobin
- WeightedRandom
- WeightRoundRobin
## How to use

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

## DynamicWeighted Loadbalance Parallel Benchmark Result

`-3`  indicate that there are 3 instances in `Selector`

`-16384 ` indicate that there are 16384 instances in `Selector`

- WeightDoubleQueue Benchmark Result

```
goos: linux
goarch: amd64
pkg: github.com/ydmxcz/loadbalance
cpu: Intel(R) Core(TM) i5-1035G1 CPU @ 1.00GHz
=== RUN   Benchmark_DynamicWeighted_Parallel
Benchmark_DynamicWeighted_Parallel
=== RUN   Benchmark_DynamicWeighted_Parallel/DynamicWeighted_3_Instances
Benchmark_DynamicWeighted_Parallel/DynamicWeighted_3_Instances
Benchmark_DynamicWeighted_Parallel/DynamicWeighted_3_Instances-8
24744768                47.46 ns/op            0 B/op          0 allocs/op
=== RUN   Benchmark_DynamicWeighted_Parallel/DynamicWeighted_16384_Instances
Benchmark_DynamicWeighted_Parallel/DynamicWeighted_16384_Instances
Benchmark_DynamicWeighted_Parallel/DynamicWeighted_16384_Instances-8
23361574                50.54 ns/op            0 B/op          0 allocs/op
PASS
ok      github.com/ydmxcz/loadbalance   2.925s
```


- Other Load-Balance Implementations benchmark results

```plaintext
goos: linux
goarch: amd64
pkg: github.com/ydmxcz/loadbalance
cpu: Intel(R) Core(TM) i5-1035G1 CPU @ 1.00GHz
=== RUN   Benchmark_GetterLoadBalance_Get_Parallel
Benchmark_GetterLoadBalance_Get_Parallel
=== RUN   Benchmark_GetterLoadBalance_Get_Parallel/Random_LoadBalance-3
Benchmark_GetterLoadBalance_Get_Parallel/Random_LoadBalance-3
Benchmark_GetterLoadBalance_Get_Parallel/Random_LoadBalance-3-8
20297527                56.63 ns/op            0 B/op          0 allocs/op
=== RUN   Benchmark_GetterLoadBalance_Get_Parallel/Random_LoadBalance-16384
Benchmark_GetterLoadBalance_Get_Parallel/Random_LoadBalance-16384
Benchmark_GetterLoadBalance_Get_Parallel/Random_LoadBalance-16384-8
18015231                57.05 ns/op            0 B/op          0 allocs/op
=== RUN   Benchmark_GetterLoadBalance_Get_Parallel/RoundRobin_LoadBalance-3
Benchmark_GetterLoadBalance_Get_Parallel/RoundRobin_LoadBalance-3
Benchmark_GetterLoadBalance_Get_Parallel/RoundRobin_LoadBalance-3-8
18871276                63.51 ns/op            0 B/op          0 allocs/op
=== RUN   Benchmark_GetterLoadBalance_Get_Parallel/RoundRobin_LoadBalance-16384
Benchmark_GetterLoadBalance_Get_Parallel/RoundRobin_LoadBalance-16384
Benchmark_GetterLoadBalance_Get_Parallel/RoundRobin_LoadBalance-16384-8
19355181                62.16 ns/op            0 B/op          0 allocs/op
=== RUN   Benchmark_GetterLoadBalance_Get_Parallel/WeightRoundRobin_3_Instances
Benchmark_GetterLoadBalance_Get_Parallel/WeightRoundRobin_3_Instances
Benchmark_GetterLoadBalance_Get_Parallel/WeightRoundRobin_3_Instances-8
22504696                52.33 ns/op            0 B/op          0 allocs/op
=== RUN   Benchmark_GetterLoadBalance_Get_Parallel/WeightRoundRobin_16384_Instances
Benchmark_GetterLoadBalance_Get_Parallel/WeightRoundRobin_16384_Instances
Benchmark_GetterLoadBalance_Get_Parallel/WeightRoundRobin_16384_Instances-8
   33603             35602 ns/op               0 B/op          0 allocs/op
=== RUN   Benchmark_GetterLoadBalance_Get_Parallel/WeightedRandom_LoadBalance-3
Benchmark_GetterLoadBalance_Get_Parallel/WeightedRandom_LoadBalance-3
Benchmark_GetterLoadBalance_Get_Parallel/WeightedRandom_LoadBalance-3-8
20667918                58.46 ns/op            0 B/op          0 allocs/op
=== RUN   Benchmark_GetterLoadBalance_Get_Parallel/WeightedRandom_LoadBalance-16384
Benchmark_GetterLoadBalance_Get_Parallel/WeightedRandom_LoadBalance-16384
Benchmark_GetterLoadBalance_Get_Parallel/WeightedRandom_LoadBalance-16384-8
  159358              7232 ns/op               0 B/op          0 allocs/op
PASS
ok      github.com/ydmxcz/loadbalance   11.602s
```
