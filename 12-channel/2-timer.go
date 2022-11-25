package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan int)
	quit := make(chan bool)

	go func() {
		for {
			select {
			case num := <-ch: //如果有数据，下面打印。但是有可能ch一直没数据
				fmt.Println("received num = ", num)
			case <-time.After(3 * time.Second): //上面的ch如果一直没数据会阻塞，那么select也会检测其他case条件，检测到后3秒超时
				fmt.Println("TimeOut")
				quit <- true
			}
		}
	}()
	for i := 0; i < 3; i++ {
		ch <- i
		time.Sleep(time.Second)
	}
	<-quit //这里暂时阻塞，直到可读
	fmt.Println("Over")
}

/* Analysis：实现超时控制
等待若干秒后，如果 ch 还没有读出数据或者被关闭，就直接结束

*/
