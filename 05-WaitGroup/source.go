package main

import (
	"sync/atomic"
	"unsafe"
)

func main() {
	// 源码地址: src/sync/waitgroup.go
}

type WaitGroup struct {
	noCopy noCopy

	// 64位数值：高32位是计数器，低32位是服务器计数。
	// 64位原子操作需要64位对齐，但32位编译器只保证64位字段是32位对齐的。
	// 由于这个原因，在32位架构上我们需要在state()中检查state1是否对齐，
	// 并在需要时动态地 "交换 "字段的顺序。
	state1 uint64
	state2 uint32
}

/*
	noCopy 字段意义：

通过给 WaitGroup 添加一个 noCopy 字段，我们就可以为 WaitGroup 实现 Locker 接口，
这样 vet 工具就可以做复制检查了。而且因为 noCopy 字段是未输出类型，
所以 WaitGroup 不会暴露 Lock/Unlock 方法。
*/
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

// state返回存储在wg.state*中的状态和sema字段的指针。
func (wg *WaitGroup) state() (statep *uint64, semap *uint32) {
	if unsafe.Alignof(wg.state1) == 8 || uintptr(unsafe.Pointer(&wg.state1))%8 == 0 {
		// state1是64位对齐的：没什么可做的。
		return &wg.state1, &wg.state2
	} else {
		// state1是32位对齐的，但不是64位对齐的：这意味着 (&state1)+4是64位对齐的。
		state := (*[3]uint32)(unsafe.Pointer(&wg.state1))
		return (*uint64)(unsafe.Pointer(&state[1])), &state[0]
	}
}

/*
	Add 方法的逻辑:

Add 方法主要操作的是 state 的计数部分。
你可以为计数值增加一个 delta 值，内部通过原子操作把这个值加到计数值上。
需要注意的是，这个 delta 也可以是个负数，相当于为计数值减去一个值，
Done 方法内部其实就是通过 Add(-1) 实现的。
*/
func (wg *WaitGroup) Add(delta int) {
	statep, semap := wg.state()
	// 高32bit是计数值v，所以把delta左移32，增加到计数上
	state := atomic.AddUint64(statep, uint64(delta)<<32)
	v := int32(state >> 32) // 当前计数值
	w := uint32(state)      // waiter count

	if v > 0 || w == 0 {
		return
	}

	// 如果计数值v为0并且waiter的数量w不为0，那么state的值就是waiter的数量
	// 将waiter的数量设置为0，因为计数值v也是0,所以它们俩的组合*statep直接设置为0即可。此时需要并唤醒所有的waiter
	*statep = 0
	for ; w != 0; w-- {
		runtime_Semrelease(semap, false, 0)
	}
}

// Done方法实际就是计数器减1
func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

/*
	Wait 方法的实现逻辑是：

不断检查 state 的值。如果其中的计数值变为了 0，那么说明所有的任务已完成，
调用者不必再等待，直接返回。如果计数值大于 0，说明此时还有任务没完成，
那么调用者就变成了等待者，需要加入 waiter 队列，并且阻塞住自己。
*/
func (wg *WaitGroup) Wait() {
	statep, semap := wg.state()

	for {
		state := atomic.LoadUint64(statep)
		v := int32(state >> 32) // 当前计数值
		w := uint32(state)      // waiter的数量
		if v == 0 {
			// 如果计数值为0, 调用这个方法的goroutine不必再等待，继续执行它后面的逻辑即可
			return
		}
		// 否则把waiter数量加1。期间可能有并发调用Wait的情况，所以最外层使用了一个for循环
		if atomic.CompareAndSwapUint64(statep, state, state+1) {
			// 阻塞休眠等待
			runtime_Semacquire(semap)
			// 被唤醒，不再阻塞，返回
			return
		}
	}
}
