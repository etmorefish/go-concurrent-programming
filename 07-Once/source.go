package main

import "sync/atomic"

func main() {
	// 源码地址：src/sync/once.go
}

type Once struct {
	done uint32
	m    Mutex
}

func (o *Once) Do(f func()) {
	if atomic.LoadUint32(&o.done) == 0 {
		o.doSlow(f)
	}
}

func (o *Once) doSlow(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	// 双检查
	if o.done == 0 {
		defer atomic.StoreUint32(&o.done, 1)
		f()
	}
}

/*
一个正确的 Once 实现要使用一个互斥锁，这样初始化的时候
如果有并发的 goroutine，就会进入doSlow 方法。
互斥锁的机制保证只有一个 goroutine 进行初始化，
同时利用双检查的机制（double-checking），再次判断 o.done
是否为 0，如果为 0，则是第一次执行，执行完毕后，
就将 o.done 设置为 1，然后释放锁。即使此时有多个 goroutine
 同时进入了 doSlow 方法，因为双检查的机制，后续的 goroutine
 会看到 o.done 的值为 1，也不会再次执行 f。这样既保证了并发的
 goroutine 会等待 f 完成，而且还不会多次执行 f。
*/
