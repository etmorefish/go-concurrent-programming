package main

import (
	"fmt"
	"sync"
)

func foo(l *sync.RWMutex) {
	fmt.Println("in foo")
	l.Lock()
	bar(l)
	l.Unlock()
}

func bar(l *sync.RWMutex) {
	l.Lock()
	fmt.Println("in bar")
	l.Unlock()
}

func main() {
	l := &sync.RWMutex{}
	foo(l)
}
