package main

import "context"

type CyclicBarrier interface {
	// 等待所有的参与者到达，如果被ctx.Done()中断，会返回ErrBrokenBarrier
	Await(ctx context.Context) error
	// 重置循环栅栏到初始化状态。如果当前有等待者，那么它们会返回ErrBrokenBarrier
	Reset()
	// 返回当前等待者的数量
	GetNumberWaiting() int
	// 参与者的数量
	GetParties() int
	// 循环栅栏是否处于中断状态
	IsBroken() bool
}

/*
循环栅栏的使用也很简单。循环栅栏的参与者只需调用 Await 等待，等所有的参与者都到达后，再执行下一步。
当执行下一步的时候，循环栅栏的状态又恢复到初始的状态了，可以迎接下一轮同样多的参与者。
*/
