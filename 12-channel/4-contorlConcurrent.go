package main

var limit = make(chan int, 3)

func main() {
	// …………
	for _, w := range work {
		go func() {
			limit <- 1
			w()  // if panic: use defer
			<-limit
		}()
	}
	// …………
}

/* Analysis:
构建一个缓冲型的 channel，容量为 3。接着遍历任务列表，每个任务启动一个 goroutine 去完成。
真正执行任务，访问第三方的动作在 w() 中完成，在执行 w() 之前，
先要从 limit 中拿“许可证”，拿到许可证之后，才能执行 w()，
并且在执行完任务，要将“许可证”归还。这样就可以控制同时运行的 goroutine 数。

limit <- 1, 如果在外层，就是控制系统 goroutine 的数量，可能会阻塞 for 循环，影响业务逻辑。
*/
