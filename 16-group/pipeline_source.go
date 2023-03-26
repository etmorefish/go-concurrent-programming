
package main

import (

    "golang.org/x/sync/errgroup"
)
/* 任务执行流水线 Pipeline
Go 官方文档中还提供了一个 pipeline 的例子。
这个例子是说，由一个子任务遍历文件夹下的文件，
然后把遍历出的文件交给 20 个 goroutine，
让这些 goroutine 并行计算文件的 md5。

我来把这个例子简化一下
*/
// 一个多阶段的pipeline.使用有限的goroutine计算每个文件的md5值.
func main() {
    m, err := MD5All(context.Background(), ".")
    if err != nil {
        log.Fatal(err)
    }

    for k, sum := range m {
        fmt.Printf("%s:\t%x\n", k, sum)
    }
}

type result struct {
    path string
    sum  [md5.Size]byte
}

// 遍历根目录下所有的文件和子文件夹,计算它们的md5的值.
func MD5All(ctx context.Context, root string) (map[string][md5.Size]byte, error) {
    g, ctx := errgroup.WithContext(ctx)
    paths := make(chan string) // 文件路径channel

    g.Go(func() error {
        defer close(paths) // 遍历完关闭paths chan
        return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
            ...... //将文件路径放入到paths
            return nil
        })
    })

    // 启动20个goroutine执行计算md5的任务，计算的文件由上一阶段的文件遍历子任务生成.
    c := make(chan result)
    const numDigesters = 20
    for i := 0; i < numDigesters; i++ {
        g.Go(func() error {
            for path := range paths { // 遍历直到paths chan被关闭
                ...... // 计算path的md5值，放入到c中
            }
            return nil
        })
    }
    go func() {
        g.Wait() // 20个goroutine以及遍历文件的goroutine都执行完
        close(c) // 关闭收集结果的chan
    }()


    m := make(map[string][md5.Size]byte)
    for r := range c { // 将md5结果从chan中读取到map中,直到c被关闭才退出
        m[r.path] = r.sum
    }

    // 再次调用Wait，依然可以得到group的error信息
    if err := g.Wait(); err != nil {
        return nil, err
    }
    return m, nil
}