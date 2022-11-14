package main

import "fmt"

func main() {
	var m = make(map[string]int)
	m["a"] = 0
	fmt.Printf("a=%d; b=%d\n", m["a"], m["b"])

	av, aexisted := m["a"]
	bv, bexisted := m["b"]
	fmt.Printf("a=%d, existed: %t; b=%d, existed: %t\n", av, aexisted, bv, bexisted)
}

/* Analysis of causes:
如果获取一个不存在的 key 对应的值时，会返回零值。
为了区分真正的零值和 key 不存在这两种情况，可以根据第二个返回值来区分
*/
