package main

import "fmt"

// 一个比较粗糙的检查channel 关闭的方式
func IsClosed(ch <-chan any) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}
func main() {
	c := make(chan any)
	fmt.Println(IsClosed(c)) //false
	close(c)
	fmt.Println(IsClosed(c)) //true
}


/* Analysis：
看一下代码，其实存在很多问题。
首先，IsClosed 函数是一个有副作用的函数。每调用一次，都会读出 channel 里的一个元素，改变了 channel 的状态。

其次，IsClosed 函数返回的结果仅代表调用那个瞬间，并不能保证调用之后会不会有其他 goroutine 对它进行了一些操作，
改变了它的这种状态。
*/