package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

func main() {
	c := sync.NewCond(&sync.Mutex{})
	var ready int

	for i := 0; i < 10; i++ {
		go func(i int) {
			time.Sleep(time.Duration(rand.Int63n(10)) * time.Second)

			// 加锁更改等待条件
			c.L.Lock()
			ready++
			c.L.Unlock()

			log.Printf("运动员#%d 已准备就绪\n", i)
			// 广播唤醒所有的等待者
			c.Broadcast()
		}(i)
	}

	c.L.Lock()
	// for ready != 10 {
	c.Wait()
	log.Println("裁判员被唤醒一次")
	// }
	c.L.Unlock()

	//所有的运动员是否就绪
	log.Println("所有运动员都准备就绪。比赛开始，3，2，1, ......")
}

/*
	原因在于，每一个运动员准备好之后都会唤醒所有的等待者，
	也就是这里的裁判员，比如第一个运动员准备好后就唤醒了裁判员，
	结果这个裁判员傻傻地没做任何检查，以为所有的运动员都准备好了，就继续执行了。
*/
