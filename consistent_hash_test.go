package loadbalance

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"
	"time"
	"unsafe"
)

// 长度为62
var bytes []byte = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890")

func init() {
	// 保证每次生成的随机数不一样
	rand.Seed(time.Now().UnixNano())
}

// 方法一
func RandStr1(n int) string {
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = bytes[rand.Int31()%62]
	}
	return *(*string)(unsafe.Pointer(&result))
}

// 方法二
func RandStr2(n int) string {
	result := make([]byte, n)
	rand.Read(result)
	return hex.EncodeToString(result)
}

// 方法二
func RandStr3(n int) string {
	result := make([]byte, n)
	r := XorShift64{state: uint64(time.Now().UnixNano())}
	for i := 0; i < n; i++ {
		result[i] = bytes[r.Int63()%62]
	}
	return *(*string)(unsafe.Pointer(&result))
	//rand.Read(result)
	//return hex.EncodeToString(result)
}

// 对比一下两种方法的性能
func Benchmark1(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			RandStr1(12)
		}
	})
	// 结果：539.1 ns/op
}

func Benchmark2(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			RandStr2(12)
		}
	})
	// 结果： 157.2 ns/op
}

func Benchmark3(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			RandStr3(12)
		}
	})
	// 结果： 157.2 ns/op
}

func TestOne(t *testing.T) {
	fmt.Println("方法一生成12位随机字符串: ", RandStr1(12))
	fmt.Println("方法二生成12位随机字符串: ", RandStr2(12))
	fmt.Println("方法三生成12位随机字符串: ", RandStr3(12))
}
