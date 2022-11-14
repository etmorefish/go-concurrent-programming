package main

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

func main() {

}

// 方案一：goroutine id
// 这个方案的关键第一步是获取 goroutine id，方式有两种，分别是简单方式和 hacker 方式。

/* 通过 runtime.Stack 方法获取栈帧信息，栈帧信息里包含 goroutine id
goroutine 1 [running]:
main.main()
		....../main.go:19 +0xb1
*/

// 只要解析出这个 id 即可
func GoID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// 得到id字符串
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

/* hacker 的方式
第一步：我们先获取到 TLS 对象；
第二步：再从 TLS 中获取 goroutine 结构的 g 指针；
第三步：再从 g 指针中取出 goroutine id。
*/

// 获取 goroutine id，我们就实现一个可以使用的可重入锁
// RecursiveMutex 包装一个Mutex,实现可重入
type RecursiveMutex struct {
	sync.Mutex
	owner     int64 // 当前持有锁的goroutine id
	recursion int32 // 这个goroutine 重入的次数
}

func (m *RecursiveMutex) Lock() {
	gid := goid.Get()
	// 如果当前持有锁的goroutine就是这次调用的goroutine,说明是重入
	if atomic.LoadInt64(&m.owner) == gid {
		m.recursion++
		return
	}
	m.Mutex.Lock()
	// 获得锁的goroutine第一次调用，记录下它的goroutine id,调用次数加1
	atomic.StoreInt64(&m.owner, gid)
	m.recursion = 1
}

func (m *RecursiveMutex) Unlock() {
	gid := goid.Get()
	// 非持有锁的goroutine尝试释放锁，错误的使用
	if atomic.LoadInt64(&m.owner) != gid {
		panic(fmt.Sprintf("wrong the owner(%d): %d!", m.owner, gid))
	}
	// 调用次数减1
	m.recursion--
	if m.recursion != 0 { // 如果这个goroutine还没有完全释放，则直接返回
		return
	}
	// 此goroutine最后一次调用，需要释放锁
	atomic.StoreInt64(&m.owner, -1)
	m.Mutex.Unlock()
}



// 方案二：token
// Token方式的递归锁
type TokenRecursiveMutex struct {
    sync.Mutex
    token     int64
    recursion int32
}

// 请求锁，需要传入token
func (m *TokenRecursiveMutex) Lock(token int64) {
    if atomic.LoadInt64(&m.token) == token { //如果传入的token和持有锁的token一致，说明是递归调用
        m.recursion++
        return
    }
    m.Mutex.Lock() // 传入的token不一致，说明不是递归调用
    // 抢到锁之后记录这个token
    atomic.StoreInt64(&m.token, token)
    m.recursion = 1
}

// 释放锁
func (m *TokenRecursiveMutex) Unlock(token int64) {
    if atomic.LoadInt64(&m.token) != token { // 释放其它token持有的锁
        panic(fmt.Sprintf("wrong the owner(%d): %d!", m.token, token))
    }
    m.recursion-- // 当前持有这个锁的token释放锁
    if m.recursion != 0 { // 还没有回退到最初的递归调用
        return
    }
    atomic.StoreInt64(&m.token, 0) // 没有递归调用了，释放锁
    m.Mutex.Unlock()
}
