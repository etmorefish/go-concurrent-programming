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

	// c.L.Lock()
	for ready != 10 {
		c.Wait()
		log.Println("裁判员被唤醒一次")
	}
	// c.L.Unlock()

	//所有的运动员是否就绪
	log.Println("所有运动员都准备就绪。比赛开始，3，2，1, ......")
}

/*
	出现这个问题的原因在于，cond.Wait 方法的实现是，
	把当前调用者加入到 notify 队列之中后会释放锁（如果不释放锁，
	其他 Wait 的调用者就没有机会加入到 notify 队列中了），然后一直等待；
	等调用者被唤醒之后，又会去争抢这把锁。如果调用 Wait 之前不加锁的话，
	就有可能 Unlock 一个未加锁的 Locker。所以切记，调用 cond.Wait 方法之前一定要加锁。
*/
