package main

import (
	"net"
	"sync"
	"time"
)

// 使用互斥锁保证线程(goroutine)安全
var connMu sync.Mutex
var conn net.Conn

func getConn() net.Conn {
	connMu.Lock()
	defer connMu.Unlock()

	// 返回已创建好的连接
	if conn != nil {
		return conn
	}

	// 创建连接
	conn, _ = net.DialTimeout("tcp", "baidu.com:80", 10*time.Second)
	return conn
}

// 使用连接
func main() {
	conn := getConn()
	if conn == nil {
		panic("conn is nil")
	}
}

/* 问题：
	这种方式虽然实现起来简单，但是有性能问题。
	一旦连接创建好，每次请求的时候还是得竞争锁才能读取到这个连接，
	这是比较浪费资源的，因为连接如果创建好之后，其实就不需要锁的保护了。
*/
