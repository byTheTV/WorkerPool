package main

import "context"


// Command types for managing workerks
type CommandType int

type Command struct{
	Type CommandType 
	WorkerID int	
}

type WorkerPool struct {
	jobQueue chan string	// Channel for jobs
	commandChan chan Command // Channel for worker Commands
	ctx		context.Context // Main ctx
	cancel	context.CancelFunc //Cacnel func for main ctx
}


func NewWorkerPool(ctx context.Context) *WorkerPool {
	poolctx, cancel := context.WithCancel(ctx)
	return &WorkerPool{
		jobQueue: make(chan string),
		commandChan: make(chan Command),
		ctx: poolctx,
		cancel: cancel,
	}
}


func main() {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()

}