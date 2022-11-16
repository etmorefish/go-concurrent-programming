package main

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()
	process(ctx)

	ctx = context.WithValue(ctx, "traceId", "xxml-2022")
	process(ctx)
}

func process(ctx context.Context) {
	traceId, ok := ctx.Value("traceId").(string)
	if ok {
		fmt.Printf("process over. trace_id=%s\n", traceId)
	} else {
		fmt.Printf("process over. no trace_id\n")
	}
}

// 传递共享的数据
/* Analysis:
第一次调用 process 函数时，ctx 是一个空的 context，自然取不出来 traceId。
第二次，通过 WithValue 函数创建了一个 context，并赋上了 traceId 这个 key，
自然就能取出来传入的 value 值。

*/
