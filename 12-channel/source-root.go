package main

import (
	"fmt"
	"time"
)

type user struct {
	name string
	age  int8
}

var u = user{name: "Messi", age: 35}
var g = &u

func modifyUser(pu *user) {
	fmt.Println("modifyUser Received Vaule", pu)
	pu.name = "Cristiano Ronaldo"
}

func printUser(u <-chan *user) {
	time.Sleep(2 * time.Second)
	fmt.Println("printUser goRoutine called", <-u)
}

func main() {
	c := make(chan *user, 5)
	c <- g
	fmt.Println(g)
	// modify g
	g = &user{name: "Neymar", age: 30}
	go printUser(c)
	go modifyUser(g)
	time.Sleep(5 * time.Second)
	fmt.Println(g)
}

/*Analysis：
一开始构造一个结构体 u，接着把 &u 赋值给指针 g，它的内容就是一个地址，指向 u。
main 程序里，先把 g 发送到 c，根据 copy value 的本质，进入到 chan buf 里的就是 u的地址，
它是指针 g 的值（不是它指向的内容），所以打印从 channel 接收到的元素时，
它就是 &{Messi 35}。因此，这里并不是将指针 g “发送” 到了 channel 里，只是拷贝它的值而已。
*/

/*Channel 发送和接收元素的本质是什么？

All transfer of value on the go channels happens with the copy of value.

就是说 channel 的发送和接收操作本质上都是 “值的拷贝”，无论是从 sender goroutine
的栈到 chan buf，还是从 chan buf 到 receiver goroutine，
或者是直接从 sender goroutine 到 receiver goroutine。
*/
