package main

import (
	"sync"
	"time"
)

func main() {
	var counter Counter3
	for i := 0; i < 10; i++ {
		go func() {
			for {
				counter.Count()
				time.Sleep(time.Millisecond)
			}
		}()
	}
	for {
		counter.Incr()
		time.Sleep(time.Second)
	}
}

// 线程安全的计数器类型
type Counter3 struct {
	mu    sync.RWMutex
	count uint64
}

// +1 的方法，内部使用互斥锁保护
func (c *Counter3) Incr() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

// 计数器得到的值 需要保护
func (c *Counter3) Count() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.count
}
