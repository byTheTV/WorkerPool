package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
	"os"
	"os/signal"
	"syscall"
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
	jobQueue  chan Job // Channel for jobs
	processor JobProcessor
	workerIDs map[int]struct{}   // map for active workers
	workerCtx  map[int]context.CancelFunc // Мапа для хранения функций отмены контекста рабочих
	ctx       context.Context    // Main ctx
	cancel    context.CancelFunc //Cacnel func for main ctx
	nextID    int
	mu        sync.Mutex
	wg        sync.WaitGroup
}

func NewWorkerPool(ctx context.Context, processor JobProcessor) *WorkerPool {
	poolctx, cancel := context.WithCancel(ctx)
	return &WorkerPool{
		jobQueue:  make(chan Job, 100),
		processor: processor,
		ctx:       poolctx,
		cancel:    cancel,
		workerIDs: make(map[int]struct{}),
		workerCtx:  make(map[int]context.CancelFunc),
	}
}

func (wp *WorkerPool) AddWorker() {
	wp.mu.Lock()
	id := wp.nextID
	wp.nextID++
	wp.workerIDs[id] = struct{}{}
	workerCtx, cancel := context.WithCancel(wp.ctx)
	wp.workerCtx[id] = cancel
	wp.mu.Unlock()

	wp.wg.Add(1)
	go func(workerID int) {
		defer wp.wg.Done()
		defer log.Printf("Worker %d stopped", workerID)

		for {
			select {
			case job := <-wp.jobQueue:
				wp.processor.Process(job, workerID)
			case <-workerCtx.Done():
				return
			case <-wp.ctx.Done():
				return
			}
		}
	}(id)
	log.Printf("Worker %d added", id)
}


func (wp *WorkerPool) RemoveWorker(id int) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if _, exists := wp.workerIDs[id]; exists {
		if cancel, exists := wp.workerCtx[id]; exists {
			cancel() // Завершаем контекст рабочего
			delete(wp.workerCtx, id)
			delete(wp.workerIDs, id)
			log.Printf("Worker %d removed", id)
		}
	}
}

func (wp *WorkerPool) SubmitJob(job Job) {
	select {
	case wp.jobQueue <- job:
	case <-wp.ctx.Done():
	}
}

func (wp *WorkerPool) WorkerCount() int {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return len(wp.workerIDs)
}

func (wp *WorkerPool) Stop() {
	wp.cancel()
	wp.wg.Wait()
	close(wp.jobQueue)
}
// WorkerPoolMonitor мониторит состояние пула
type WorkerPoolMonitor struct {
	pool *WorkerPool
}

func (m WorkerPoolMonitor) Report() string {
	return fmt.Sprintf("Workers active: %d", m.pool.WorkerCount())
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	processor := ConsoleProcessor{}

	pool := NewWorkerPool(ctx, processor)

	for i := 0; i < 3; i++ {
		pool.AddWorker()
	}

	// Настраиваем мониторинг
	monitor := WorkerPoolMonitor{pool: pool}
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Println(monitor.Report())
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		jobID :=1
		for {
			select {
			case <-ctx.Done():
				return
			default:
				pool.SubmitJob(Job{
					ID:   jobID,
					Data: fmt.Sprintf("Data %d", jobID),
				})
				jobID++
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	// Ручное управление рабочими для демонстрации
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if rand.Intn(2) == 0 && pool.WorkerCount() > 1 {
					
					count := pool.WorkerCount()
					for id := range pool.workerIDs {
						if id == count-1 {
							pool.RemoveWorker(id)
							break
						}
					}
				} else {
					pool.AddWorker()
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Обработка сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	pool.Stop()
	log.Println("Worker pool stopped")
}