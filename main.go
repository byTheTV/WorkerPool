package main

import "context"

type WorkerPool struct {
	ctx		context.Context // Main ctx
	cancel	context.CancelFunc //Cacnel func for main ctx
}

func NewWorkerPool(ctx context.Context) *WorkerPool {
	poolctx, cancel := context.WithCancel(ctx)
	return &WorkerPool{
		ctx: poolctx,
		cancel: cancel,
	}
}


func main() {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()

}