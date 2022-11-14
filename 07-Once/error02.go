package main

import (
	"io"
	"net"
	"os"
	"sync"
)

func main() {
	var once sync.Once
	var googleConn net.Conn // 到Google网站的一个连接

	once.Do(func() {
		// 建立到google.com的连接，有可能因为网络的原因，googleConn并没有建立成功，此时它的值为nil
		googleConn, _ = net.Dial("tcp", "google.com:80")
	})
	// 发送http请求
	googleConn.Write([]byte("GET / HTTP/1.1\r\nHost: google.com\r\n Accept: */*\r\n\r\n"))
	io.Copy(os.Stdout, googleConn)
}

/*
如果 f 方法执行的时候 panic，或者 f 执行初始化资源的时候失败了，
这个时候，Once 还是会认为初次执行已经成功了，即使再次调用 Do 方法，也不会再次执行 f。
*/
