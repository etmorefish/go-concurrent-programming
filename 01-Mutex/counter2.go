package main

import (
	"fmt"
	"sync"
)

func main() {
	//封装好计数器
	var counter Counter2

	// 并发控制任务编排
	var wg sync.WaitGroup

	wg.Add(10)
	//启动10个goroutine
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				counter.Incr()

			}
		}()
	}
	wg.Wait()
	fmt.Println(counter.Count())
}

// 线程安全计数器类型
type Counter2 struct {
	mu    sync.Mutex
	count uint64
}

// +1 方法
func (c *Counter2) Incr() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

// 得到计数器的值，也需要锁保护
func (c *Counter2) Count() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}
