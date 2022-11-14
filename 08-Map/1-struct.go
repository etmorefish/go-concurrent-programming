package main

import "fmt"

type mapKey struct {
	key int
}

func main() {
	var m = make(map[mapKey]string)
	var key = mapKey{10}

	m[key] = "hello"
	fmt.Printf("m[key]=%s\n", m[key])

	// 修改key的字段的值后再次查询map，无法获取刚才add进去的值
	key.key = 100
	fmt.Printf("再次查询m[key]=%s\n", m[key])
}

/* Notice:
这里有一点需要注意，如果使用 struct 类型做 key 其实是有坑的，
因为如果 struct 的某个字段值修改了，查询 map 时无法获取它 add 进去的值，

Solve:
如果要使用 struct 作为 key，我们要保证 struct 对象在逻辑上是不可变的，
这样才会保证 map 的逻辑没有问题。
*/
