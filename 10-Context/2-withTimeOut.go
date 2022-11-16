package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second) // 修改此处超时时间，可打印出不同结果
	defer cancel()                                                          // 避免其他地方忘记cancel，且重复调用不影响

	ids := fetchWebData(ctx)

	fmt.Println(ids)
}

func fetchWebData(ctx context.Context) (res []int64) {
	select {
	case <-time.After(3 * time.Second):
		return []int64{100, 200, 300}
	case <-ctx.Done():
		return []int64{1, 2, 3}
	}
}

// 定时取消
/*  Analysis:

注意一个细节，WithTimeOut 函数返回的 context 和 cancelFun 是分开的。
context 本身并没有取消函数，这样做的原因是取消函数只能由外层函数调用，
防止子节点 context 调用取消函数，从而严格控制信息的流向：由父节点 context 流向子节点 context。
*/
