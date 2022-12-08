package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"golang.org/x/sync/semaphore"
)

var (
	maxWorkers = runtime.GOMAXPROCS(0)                    // worker数量
	sema       = semaphore.NewWeighted(int64(maxWorkers)) //信号量
	task       = make([]int, maxWorkers*4)                // 任务数，是worker的四倍
)

func main() {

	ctx := context.Background()

	for i := range task {
		// 如果没有worker可用，会阻塞在这里，直到某个worker被释放
		if err := sema.Acquire(ctx, 1); err != nil {
			break
		}

		// 启动worker goroutine
		go func(i int) {
			defer sema.Release(1)
			time.Sleep(100 * time.Millisecond) // 模拟一个耗时操作
			task[i] = i + 1
		}(i)
	}

	// 请求所有的worker,这样能确保前面的worker都执行完
	if err := sema.Acquire(ctx, int64(maxWorkers)); err != nil {
		log.Printf("获取所有的worker失败: %v", err)
	}

	fmt.Println(task)
}

/*Analysis:
我们创建和 CPU 核数一样多的 Worker，让它们去处理一个 4 倍数量的整数 slice。
每个 Worker 一次只能处理一个整数，处理完之后，才能处理下一个。

在这段代码中，main goroutine 相当于一个 dispatcher，负责任务的分发。
它先请求信号量，如果获取成功，就会启动一个 goroutine 去处理计算，然后，
这个 goroutine 会释放这个信号量（信号量的获取是在 main goroutine，
信号量的释放是在 worker goroutine 中），如果获取不成功，就等到有信号
量可以使用的时候，再去获取。
*/
