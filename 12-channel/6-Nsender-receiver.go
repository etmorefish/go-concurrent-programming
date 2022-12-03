package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	const Max = 100000
	const NumSenders = 10

	dataCh := make(chan int, 100)
	stopCh := make(chan struct{})

	// senders
	for i := 0; i < NumSenders; i++ {
		go func() {
			for {
				select {
				case <-stopCh:
					return
				case dataCh <- rand.Intn(Max):
				}
			}
		}()
	}

	// the receiver
	go func() {
		for value := range dataCh {
			if value == Max-1 {
				fmt.Println("send stop signal to senders.")
				close(stopCh)
				return
			}

			fmt.Println(value)
		}
	}()

	select {
	case <-time.After(time.Second * 10):
	}
}

/*Analysis：
这里的 stopCh 就是信号 channel，它本身只有一个 sender，因此可以直接关闭它。
senders 收到了关闭信号后，select 分支 “case <- stopCh” 被选中，退出函数，不再发送数据。

需要说明的是，上面的代码并没有明确关闭 dataCh。
在 Go 语言中，对于一个 channel，如果最终没有任何 goroutine 引用它，
不管 channel 有没有被关闭，最终都会被 gc 回收。所以，在这种情形下，不关闭 channel，让 gc 代劳。
*/
