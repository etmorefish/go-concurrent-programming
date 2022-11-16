package main

import (
	"context"
	"fmt"
	"time"
)

func main() {

	// 这是一个可以生成无限整数的协程，但如果我只需要它产生的前 5 个数，那么就会发生 goroutine 泄漏：
	for n := range gen() {
		fmt.Println(n)
		if n == 5 {
			break
			// 当 n == 5 的时候，直接 break 掉。
			// 那么 gen 函数的协程就会执行无限循环，永远不会停下来。
			// 发生了 goroutine 泄漏。
		}
	}
	// ……

	/*
		// 增加一个 context，在 break 前调用 cancel 函数，取消 goroutine。
		// gen 函数在接收到取消信号后，直接退出，系统回收资源。
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel() // 避免其他地方忘记 cancel，且重复调用不影响

		for n := range gen2(ctx) {
			fmt.Println(n)
			if n == 5 {
				cancel()
				break
			}
		}
		// ……
	*/
}

func gen() <-chan int {
	ch := make(chan int)
	go func() {
		var n int
		for {
			ch <- n
			n++
			time.Sleep(time.Second)
		}
	}()
	return ch
}

func gen2(ctx context.Context) <-chan int {
	ch := make(chan int)
	go func() {
		var n int
		for {
			select {
			case <-ctx.Done():
				return
			case ch <- n:
				n++
				time.Sleep(time.Second)
			}
		}
	}()
	return ch
}
