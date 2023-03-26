package main

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

/*
	返回所有子任务的错误

Group 只能返回子任务的第一个错误，后续的错误都会被丢弃。
但是，有时候我们需要知道每个任务的执行情况。怎么办呢？
这个时候，我们就可以用稍微有点曲折的方式去实现。
我们使用一个 result slice 保存子任务的执行结果，
这样，通过查询 result，就可以知道每一个子任务的结果了。
*/
func main() {
	var g errgroup.Group
	var result = make([]error, 3)

	// 启动第一个子任务,它执行成功
	g.Go(func() error {
		time.Sleep(5 * time.Second)
		fmt.Println("exec #1")
		result[0] = nil // 保存成功或者失败的结果
		return nil
	})

	// 启动第二个子任务，它执行失败
	g.Go(func() error {
		time.Sleep(10 * time.Second)
		fmt.Println("exec #2")

		result[1] = errors.New("failed to exec #2") // 保存成功或者失败的结果
		return result[1]
	})

	// 启动第三个子任务，它执行成功
	g.Go(func() error {
		time.Sleep(15 * time.Second)
		fmt.Println("exec #3")
		result[2] = nil // 保存成功或者失败的结果
		return nil
	})

	if err := g.Wait(); err == nil {
		fmt.Printf("Successfully exec all. result: %v\n", result)
	} else {
		fmt.Printf("failed: %v\n", result)
	}
}

/*
就是使用 result 记录每个子任务成功或失败的结果。
其实，你不仅可以使用 result 记录 error 信息，还可以用它记录计算结果
*/
