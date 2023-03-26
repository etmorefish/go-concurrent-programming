package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/marusama/cyclicbarrier"
)

// 创建一个只允许10个参与者通过的障碍物，每次当所有进程达到障碍物时，该动作将会被执行。
func main() {
	cnt := 0
	b := cyclicbarrier.NewWithAction(10, func() error {
		cnt++
		return nil
	})

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ { // 创建10个goroutine，其数量与障碍物中的参与者数量相同。
		wg.Add(1)
		go func() {
			for j := 0; j < 5; j++ {

				// 做一些复杂的任务5次。
				time.Sleep(100 * time.Millisecond)

				err := b.Await(context.TODO()) // 等待障碍物中其他参与者完成，然后执行障碍物的动作，将所有其他进程传递给下一轮。
				if err != nil {
					panic(err)
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println(cnt) // cnt=5，它表示障碍物被推倒了5次。
}
