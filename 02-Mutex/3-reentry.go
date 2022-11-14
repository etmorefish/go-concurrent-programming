package main

import (
	"fmt"
	"sync"
)

func foo(l sync.Locker) {
	fmt.Println("in foo")
	l.Lock() //1 
	bar(l)
	l.Unlock()
}

func bar(l sync.Locker) {
	l.Lock() //2
	fmt.Println("in bar")
	l.Unlock()
}

func main() {
	l := &sync.Mutex{}
	foo(l)
}
