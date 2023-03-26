package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

type Data struct {
}

func getData() (*Data, error) {
	time.Sleep(3 * time.Second)
	return &Data{}, nil
}

func main() {
	c, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(c)

	datas := make(chan *Data, 10)

	g.Go(func() error {
		// 业务逻辑
		data, err := getData()
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		datas <- data
		return nil
	})

	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()

	err := g.Wait()
	if err == nil {
		fmt.Println("success")
		return
	}
	fmt.Println("fail", err)
}
