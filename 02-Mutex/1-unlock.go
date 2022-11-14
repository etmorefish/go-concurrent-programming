package main

import (
	"fmt"
	"sync"
)

func main() {
	foo()
}
func foo() {
	var mu sync.Mutex
	defer mu.Unlock()
	fmt.Println("hello world!")
}
