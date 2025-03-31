package scaling_worker_pool

import (
	"sync"
)

type WorkerPool[T any] struct {
	sync.Mutex
	wg       sync.WaitGroup
	workerFn func(job T)
	workCh   chan T
	stopChs  []chan struct{}
}

func NewWorkerPool[T any](workerFn func(T), queueSize int) *WorkerPool[T] {
	return &WorkerPool[T]{
		workerFn: workerFn,
		workCh:   make(chan T, queueSize),
		stopChs:  make([]chan struct{}, 0),
	}
}

func (wp *WorkerPool[T]) SetWorkerCount(n int) {
	wp.Lock()
	defer wp.Unlock()

	currWorkers := len(wp.stopChs)
	switch {
	case n > currWorkers:
		for i := currWorkers; i < n; i++ {
			stopCh := make(chan struct{})
			wp.stopChs = append(wp.stopChs, stopCh)
			wp.wg.Add(1)
			go wp.runWorker(stopCh)
		}
	case n < currWorkers:
		toStop := currWorkers - n
		for range toStop {
			last := len(wp.stopChs) - 1
			close(wp.stopChs[last])
			wp.stopChs = wp.stopChs[:last]
		}
	}
}

func (wp *WorkerPool[T]) runWorker(stopCh chan struct{}) {
	defer wp.wg.Done()
	for {
		select {
		case job := <-wp.workCh:
			wp.workerFn(job)
		case <-stopCh:
			return
		}
	}
}

func (wp *WorkerPool[T]) Send(job T) {
	wp.workCh <- job
}

func (wp *WorkerPool[T]) StopAllWorkers() {
	wp.SetWorkerCount(0)
	wp.wg.Wait()
}
