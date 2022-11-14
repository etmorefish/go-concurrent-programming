package main

import "fmt"

func main() {
	var m map[int]int
	fmt.Println(m[100])
}

// 从一个 nil 的 map 对象中获取值不会 panic，而是会得到零值，所以代码不会报错
