package main

import "sync/atomic"

func main() {
	// 源码地址： src/sync/rwmutex.go
}

type Mutex struct {
	state int32
	sema  uint32
}

type RWMutex struct {
	w           Mutex  // 互斥锁解决多个writer的竞争
	writerSem   uint32 // writer信号量
	readerSem   uint32 // reader信号量
	readerCount int32  // reader的数量
	readerWait  int32  // writer等待完成的reader的数量
}

const rwmutexMaxReaders = 1 << 30 // 定义了最大的 reader 数量

// RLock/RUnlock 的实现,移除了 race 等无关紧要的代码
func (rw *RWMutex) RLock() {
	if atomic.AddInt32(&rw.readerCount, 1) < 0 {
		// rw.readerCount是负值的时候，意味着此时有writer等待请求锁，因为writer优先级高，所以把后来的reader阻塞休眠
		runtime_SemacquireMutex(&rw.readerSem, false, 0)
	}
}
func (rw *RWMutex) RUnlock() {
	if r := atomic.AddInt32(&rw.readerCount, -1); r < 0 {
		rw.rUnlockSlow(r) // 有等待的writer
	}
}
func (rw *RWMutex) rUnlockSlow(r int32) {
	if atomic.AddInt32(&rw.readerWait, -1) == 0 {
		// 最后一个reader了，writer终于有机会获得锁了
		runtime_Semrelease(&rw.writerSem, false, 1)
	}
}

// ----------------------------------------------------------------
func (rw *RWMutex) Lock() {
	// 首先解决其他writer竞争问题
	rw.w.Lock()
	// 反转readerCount，告诉reader有writer竞争锁
	r := atomic.AddInt32(&rw.readerCount, -rwmutexMaxReaders) + rwmutexMaxReaders
	// 如果当前有reader持有锁，那么需要等待
	if r != 0 && atomic.AddInt32(&rw.readerWait, r) != 0 {
		runtime_SemacquireMutex(&rw.writerSem, false, 0)
	}
}

func (rw *RWMutex) Unlock() {
	// 告诉reader没有活跃的writer了
	r := atomic.AddInt32(&rw.readerCount, rwmutexMaxReaders)

	// 唤醒阻塞的reader们
	for i := 0; i < int(r); i++ {
		runtime_Semrelease(&rw.readerSem, false, 0)
	}
	// 释放内部的互斥锁
	rw.w.Unlock()
}
