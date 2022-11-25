package main

import (
	"fmt"
	"time"
)

func main() {
	worker()
}

func worker() {
	ticker := time.Tick(1 * time.Second)
	for {
		select {
		case <-ticker:
			// 执行定时任务
			// dosomething
			fmt.Println("执行 1s 定时任务")
		}
	}
}

/* Analysis: 定时任务
每隔 1 秒种，执行一次定时任务。
*/
