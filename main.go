package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Job struct {
	ID   int
	Data string
}

type JobProcessor interface {
	Process(job Job, workerID int)
}

// Реализует JobProcessor
type ConsoleProcessor struct{}

func (p ConsoleProcessor) Process(job Job, workerID int) {
	fmt.Printf("Worker %d: Job %d - %s\n", workerID, job.ID, job.Data)
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
}


type WorkerPool struct {
	jobQueue    chan Job        // Channel for jobs
	processor  JobProcessor
	workerIDs  map[int]struct{} // map for active workers
	ctx         context.Context    // Main ctx
	cancel      context.CancelFunc //Cacnel func for main ctx
	nextID		int
	mu         sync.Mutex
	wg          sync.WaitGroup
}

func NewWorkerPool(ctx context.Context, processor JobProcessor) *WorkerPool {
	poolctx, cancel := context.WithCancel(ctx)
	return &WorkerPool{
		jobQueue:    make(chan Job, 100),
		processor: 	processor,
		ctx:         poolctx,
		cancel:      cancel,
		workerIDs: make(map[int]struct{}),
	}
}

func(wp *WorkerPool) AddWorker(){
	wp.mu.Lock()
	id := wp.nextID
	wp.nextID++
	wp.workerIDs[id] = struct{}{}
	wp.mu.Unlock()

	wp.wg.Add(1)
	go func(workerID int) {
		defer wp.wg.Done()
		defer log.Printf("Worker %d stopped", workerID)

		for{
			select{
			case job := <- wp.jobQueue:
				wp.processor.Process(job, workerID)
			case <-wp.ctx.Done():
				return
			}
		}
	} (id)
	log.Printf("Worker %d added", id)
}

func main() {
	//	ctx, cancel := context.WithCancel(context.Background())
	//	defer cancel()

}
