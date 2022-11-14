package main

func main() {
	var m = make(map[int]int, 10) // 初始化一个map
	go func() {
		for {
			m[1] = 1 //设置key
		}
	}()

	go func() {
		for {
			_ = m[2] //访问这个map
		}
	}()
	select {}
}

/* Analysis:
虽然这段代码看起来是读写 goroutine 各自操作不同的元素，
貌似 map 也没有扩容的问题，但是运行时检测到同时对 map
对象有并发访问，就会直接 panic。panic 信息会告诉我们代
码中哪一行有读写问题，根据这个错误信息你就能快速定位出来
是哪一个 map 对象在哪里出的问题了。
*/
