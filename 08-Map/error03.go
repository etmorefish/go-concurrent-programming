package main

import (
	"fmt"
	"time"
)

type Counter struct {
	Website      string
	Start        time.Time
	PageCounters map[string]int
}

func main() {
	var c Counter
	c.Website = "baidu.com"

	// c.PageCounters["/"]++
	m := map[string]int{"p1": 1, "p2": 2}
	c.PageCounters = m

	fmt.Printf("%+v", c)
}

// panic: assignment to entry in nil map
// map 的初始化问题。
// 有时候 map 作为一个 struct 字段的时候，很容易忘记初始化
