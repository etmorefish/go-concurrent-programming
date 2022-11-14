package main

import "sync"

// 常见问题一：计数器设置为负值

// 第一种方法是：调用 Add 的时候传递一个负数
func main() {
	var wg sync.WaitGroup
	wg.Add(10)

	wg.Add(-10) //将-10作为参数调用Add，计数值被设置为0

	wg.Add(-1) //将-1作为参数调用Add，如果加上-1计数值就会变为负数。这是不对的，所以会触发panic
}

// 2 调用 Done 方法的次数过多，超过了 WaitGroup 的计数值。
// 使用 WaitGroup 的正确姿势是，预先确定好 WaitGroup 的计数值，然后调用相同次数的 Done 完成相应的任务。
func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	wg.Done()

	wg.Done() // 1-1-1=-1
}
