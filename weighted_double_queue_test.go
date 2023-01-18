package loadbalance_test

import (
	"fmt"

	"github.com/ydmxcz/loadbalance"
)

func Example() {

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
	_, _ = selector.Get("192.168.0.90:9002")

	_ = selector.Size()

	// check all instance by for-each
	selector.ForEach(func(_ string, ms *myService) bool {
		fmt.Println(ms)
		return true
	})

}
