package main

import (
	"context"
	"fmt"

	"github.com/YouDecideIt/auto-index/operations"
)

func main() {
	op := operations.New()
	ctx := context.Background()
	defer func() {
		if err := op.DestroyBCluster(ctx); err != nil {
			panic(err)
		}
	}()
	if err := op.StartBCluster(ctx); err != nil {
		panic(err)
	}
	c, err := op.WaitBClusterStartedAndMirrored(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(c)
}
