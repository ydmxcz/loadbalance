package loadbalance_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/ydmxcz/loadbalance"
)

func BenchmarkXorShift64Parallel(b *testing.B) {
	xs := loadbalance.NewXorShift64(uint64(time.Now().UnixNano()))
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			xs.Int63()
		}
	})
}

func TestRandomNum(t *testing.T) {
	// xs := XorShift64{State: uint64(time.Now().UnixNano())}
	xs := rand.New(rand.NewSource(time.Now().UnixNano()))

	// hashfunc := GetHashFunc[int64]()
	m := make(map[uint64]int)
	for i := 0; i < 100000; i++ {
		m[uint64(xs.Int63())%16]++
	}
	fmt.Println(m)
	fmt.Println(len(m))
}

func TestXorShift64(t *testing.T) {
	xs := loadbalance.NewXorShift64(uint64(time.Now().UnixNano()))

	m := make(map[uint64]int)
	for i := 0; i < 100000; i++ {
		m[uint64(xs.Int63())%16]++
	}
	fmt.Println(m)
	fmt.Println(len(m))
}

func BenchmarkRandSourceParallel(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			rand.Int63()
		}
	})
}
