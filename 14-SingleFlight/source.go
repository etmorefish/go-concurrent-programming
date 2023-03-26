package main

import "sync"

// 结构

type Group struct { // singleflight实体
	mu sync.Mutex       // 互斥锁
	m  map[string]*call // 懒加载
}

type call struct {
	wg sync.WaitGroup
	// 存储 调用singleflight.Do()方法返回的结果
	val interface{}
	err error

	// 调用singleflight.Forget(key)时将对应的key从Group.m中删除
	forgotten bool

	// 通俗的理解成singleflight合并的并发请求数
	dups int
	// 存储 调用singleflight.DoChan()方法返回的结果
	chans []chan<- Result
}

type Result struct {
	Val    interface{}
	Err    error
	Shared bool
}

// 对外暴露的方法
// 这个方法比较灵性，通过两个 defer 巧妙的区分了到底是发生了 panic 还是用户主动调用了 runtime.Goexit，逻辑还是比较复杂
func Do(key string, fn func() (interface{}, error)) (v interface{}, err error, shared bool)

// 和 do 唯一的区别是 go g.doCall(c, key, fn),但对起了一个 goroutine 来执行，
// 并通过 channel 来返回数据，这样外部可以自定义超时逻辑，防止因为 fn 的阻塞，导致大量请求都被阻塞。
func DoChan(key string, fn func() (interface{}, error)) <-chan Result

// 手动移除某个 key，让后续请求能走 doCall 的逻辑，而不是直接阻塞。
func Forget(key string)

// DoChan()和Do()最大的区别是DoChan()属于异步调用，返回一个channel，解决同步调用时的阻塞问题
