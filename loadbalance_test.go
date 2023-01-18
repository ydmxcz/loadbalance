package loadbalance_test

import (
	"fmt"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/alphadose/haxmap"
	"github.com/ydmxcz/loadbalance"
)

type myService struct {
	Address         string
	Memory          int
	currentWeight   int
	effectiveWeight int
}

func (ms *myService) InstanceID() string {
	return ms.Address
}

func (ms *myService) InstanceWeight() int {
	return ms.Memory
}

// getInstance get a slice of *MyService for test
// when the mode equals 1: return three instances,and the weight is 5,3,2.
// when the mode equals 2: return the sum of the number of instances is 16384,
// the weight of instance as the square of 2, from 2 to 128,
// see code for details about the number of instance in different weight.
func getInstance(mode int) []*myService {
	instances := make([]*myService, 0)

	addInstences := func(instances []*myService, nodeNum, nodeWeight int) []*myService {
		for i := 1; i <= nodeNum; i++ {
			instances = append(instances, &myService{
				Address: fmt.Sprintf("192.168.0.%d:%d", nodeWeight, i),
				Memory:  nodeWeight,
			})
		}
		return instances
	}
	if mode == 1 {
		instances = addInstences(instances, 1, 5)
		instances = addInstences(instances, 1, 3)
		instances = addInstences(instances, 1, 2)
	} else if mode == 2 {
		instances = addInstences(instances, 256, 128)    // sum:256
		instances = addInstences(instances, 512+256, 64) // sum:1024
		instances = addInstences(instances, 1024, 32)    // sum:2048
		instances = addInstences(instances, 2048, 16)    // sum:4096
		instances = addInstences(instances, 3072, 8)     // sum:7168
		instances = addInstences(instances, 4096, 4)     // sum:11264
		instances = addInstences(instances, 5120, 2)     // sum:16384
	}
	return instances
}

func TestSupplirLoadBalance(t *testing.T) {
	testSupplirLoadBalance("Random LoadBalance", loadbalance.NewRandom[string, *myService](), t)
	testSupplirLoadBalance("RoundRobin LoadBalance", loadbalance.NewRoundRobin[string, *myService](), t)
	testSupplirLoadBalance("WeightedDoubleQueue LoadBalance", loadbalance.NewWeightedDoubleQueue[string, *myService](), t)
	testSupplirLoadBalance("WeightRoundRobin LoadBalance", loadbalance.NewWeightRoundRobin[string, *myService](), t)
	testSupplirLoadBalance("WeightedRandom LoadBalance", loadbalance.NewWeightedRandom[string, *myService](), t)
}

func testSupplirLoadBalance(name string, lb loadbalance.Selector[string, *myService], t *testing.T) {
	// detial in comment
	l := getInstance(1)
	lb.Add(l...)

	m := map[string]int{}
	for i := 0; i < 10000; i++ {
		m[lb.Select().InstanceID()]++
	}
	fmt.Println(name, ":")
	for k, v := range m {
		fmt.Printf("[%s]::%d\n", k, v)

	}
}

func benchmarkLoadBalanceParallel(lb loadbalance.Selector[string, *myService],
	insts []*myService, b *testing.B) {
	lb.Add(insts...)
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			if lb.Select() == nil {
				b.Fatal("Nil")
			}
		}
	})
}

func Benchmark_WeightedDoubleQueue(b *testing.B) {
	b.Run("WeightedDoubleQueue LoadBalance-3", func(b *testing.B) {
		benchmarkLoadBalanceParallel(loadbalance.NewWeightedDoubleQueue[string, *myService](), getInstance(1), b)
	})
	//av
	b.Run("WeightedDoubleQueue LoadBalance-16384", func(b *testing.B) {
		benchmarkLoadBalanceParallel(loadbalance.NewWeightedDoubleQueue[string, *myService](), getInstance(2), b)
	})
}

func Benchmark_GetterLoadBalance_Get_Parallel(b *testing.B) {
	b.Run("Random LoadBalance-3", func(b *testing.B) {
		benchmarkLoadBalanceParallel(loadbalance.NewRandom[string, *myService](), getInstance(1), b)
	})
	b.Run("Random LoadBalance-16384", func(b *testing.B) {
		benchmarkLoadBalanceParallel(loadbalance.NewRandom[string, *myService](), getInstance(2), b)
	})
	b.Run("RoundRobin LoadBalance-3", func(b *testing.B) {
		benchmarkLoadBalanceParallel(loadbalance.NewRoundRobin[string, *myService](), getInstance(1), b)
	})
	b.Run("RoundRobin LoadBalance-16384", func(b *testing.B) {
		benchmarkLoadBalanceParallel(loadbalance.NewRoundRobin[string, *myService](), getInstance(2), b)
	})
	b.Run("WeightedDoubleQueue LoadBalance-3", func(b *testing.B) {
		benchmarkLoadBalanceParallel(loadbalance.NewWeightedDoubleQueue[string, *myService](), getInstance(1), b)
	})
	b.Run("WeightedDoubleQueue LoadBalance-16384", func(b *testing.B) {
		benchmarkLoadBalanceParallel(loadbalance.NewWeightedDoubleQueue[string, *myService](), getInstance(2), b)
	})
	b.Run("WeightRoundRobin LoadBalance-3", func(b *testing.B) {
		benchmarkLoadBalanceParallel(loadbalance.NewWeightRoundRobin[string, *myService](), getInstance(1), b)
	})
	b.Run("WeightRoundRobin LoadBalance-16384", func(b *testing.B) {
		benchmarkLoadBalanceParallel(loadbalance.NewWeightRoundRobin[string, *myService](), getInstance(2), b)
	})
	b.Run("WeightedRandom LoadBalance-3", func(b *testing.B) {
		benchmarkLoadBalanceParallel(loadbalance.NewWeightedRandom[string, *myService](), getInstance(1), b)
	})
	b.Run("WeightedRandom LoadBalance-16384", func(b *testing.B) {
		benchmarkLoadBalanceParallel(loadbalance.NewWeightedRandom[string, *myService](), getInstance(2), b)
	})
}

func clearSymbol(text []byte, check func(rune) bool) []byte {
	var k, i int
	for i < len(text) {
		r, s := utf8.DecodeRune(text[i:])
		if check(r) {
			utf8.EncodeRune(text[k:], r)
			k += s
		}
		i += s
	}
	return text[:k]
}

func checkSymbol(r rune) bool {
	switch r {
	case '，', '；', ',', ';', ' ', ':':
		return false
	default:
		return true
	}
}

func TestCheckSymbol(t *testing.T) {

	str := "   a b ；c， d;;      e:f,g "
	if string(clearSymbol([]byte(str), checkSymbol)) != "abcdefg" {
		t.Fatal("Result Error")
	}
}

func BenchmarkMapGet(b *testing.B) {
	ins := getInstance(2)
	m := map[string]loadbalance.Instance[string]{}
	for j := 0; j < len(ins); j++ {
		m[ins[j].InstanceID()] = ins[j]
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(ins); j++ {
			_, ok := m[ins[j].InstanceID()]
			if !ok {
				b.Fatal("NULL")
			}
		}

	}
}

func BenchmarkHaxMapGet(b *testing.B) {
	ins := getInstance(2)
	// m := map[string]loadbalance.Instance[string]{}
	m := haxmap.New[string, loadbalance.Instance[string]]() //map[string]loadbalance.Instance[string]{}
	for j := 0; j < len(ins); j++ {
		m.Set(ins[j].InstanceID(), ins[j])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(ins); j++ {
			_, ok := m.Get(ins[j].InstanceID())
			if !ok {
				b.Fatal("NULL")
			}
		}

	}
}

func BenchmarkMapGet_Parallel(b *testing.B) {
	ins := getInstance(2)
	m := map[string]loadbalance.Instance[string]{}
	for j := 0; j < len(ins); j++ {
		m[ins[j].InstanceID()] = ins[j]
	}
	mutex := sync.Mutex{}
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			for j := 0; j < len(ins); j++ {
				mutex.Lock()
				_, ok := m[ins[j].InstanceID()]
				mutex.Unlock()
				if !ok {
					b.Fatal("NULL")
				}
			}
		}
	})
}

func BenchmarkHaxMapGet_Parallel(b *testing.B) {
	ins := getInstance(2)
	// m := map[string]loadbalance.Instance[string]{}
	m := haxmap.New[string, loadbalance.Instance[string]]() //map[string]loadbalance.Instance[string]{}
	for j := 0; j < len(ins); j++ {
		m.Set(ins[j].InstanceID(), ins[j])
	}
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			for j := 0; j < len(ins); j++ {
				_, ok := m.Get(ins[j].InstanceID())
				if !ok {
					b.Fatal("NULL")
				}
			}
		}
	})
}
