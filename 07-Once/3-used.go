package main

import (
	"fmt"
	"sync"
)

func main() {
	var once sync.Once

	// 第一个初始化函数
	f1 := func() {
		fmt.Println("in f1")
	}
	once.Do(f1) // 打印出 in f1

	// 第二个初始化函数
	f2 := func() {
		fmt.Println("in f2")
	}
	once.Do(f2) // 无输出
}

/*
因为这里的 f 参数是一个无参数无返回的函数，所以你可能会通过闭包的方式引用外面的参数
*/
