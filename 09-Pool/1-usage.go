package main

import (
	"fmt"
	"sync"
)

var pool *sync.Pool

type Person struct {
	Name string
}

func initPool() {
	pool = &sync.Pool{
		New: func() interface{} {
			fmt.Println("Creating a new Person...")
			return new(Person)
		},
	}
}

func main() {
	initPool()

	p := pool.Get().(*Person)
	fmt.Println("首次从 pool 里获取： ", p)

	p.Name = "first"
	fmt.Printf("设置 p.Name = %s\n", p.Name)

	pool.Put(p)

	fmt.Println("Pool 里已有一个对象： &{first}, 调用 Get: ", pool.Get().(*Person))
	fmt.Println("Pool 没有对象了：, 调用 Get: ", pool.Get().(*Person))

}

/* Analysis:
首先，需要初始化 Pool，唯一需要做的就是设置好 New 函数。
当调用 Get 方法时，如果池子里缓存了对象，就直接返回缓存的对象。
如果没有“存货”，则调用 New 两数创建一个新的对象。

另外，Get 方法取出来的对象和上次Put 进去的对象实际上是同一个，
Pool 没有做任何“清空”的处理。但不应当对此有任何假设，
因为在实际的并发使用场景中，无法保证这种顺序，最好的做法是在_Put 前，将对象清空。

*/
