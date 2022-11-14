package main

import (
	"fmt"
	"sync"
)

func main() {
	var once sync.Once
	// once.Do(func() {
	once.Do(func() {
		fmt.Println("初始化")
	})
	// })
}

// 1. 死锁
// Do 方法会执行一次 f，但是如果 f 中再次调用这个 Once 的 Do 方法的话，
// 就会导致死锁的情况出现。这还不是无限递归的情况，而是的的确确的 Lock 的递归调用导致的死锁。
// 想要避免这种情况的出现，就不要在 f 参数中调用当前的这个 Once，不管是直接的还是间接的。
