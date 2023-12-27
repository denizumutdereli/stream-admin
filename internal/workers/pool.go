package workers

import "context"

type taskFunc func() error

type WorkerPool struct {
	taskQueue  chan taskFunc
	ErrorQueue chan error
	done       chan struct{}
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	wp := &WorkerPool{
		taskQueue:  make(chan taskFunc, numWorkers),
		ErrorQueue: make(chan error, numWorkers),
		done:       make(chan struct{}),
	}

	for i := 0; i < numWorkers; i++ {
		go wp.worker()
	}

	return wp
}

func (wp *WorkerPool) worker() {
	for {
		select {
		case task, ok := <-wp.taskQueue:
			if !ok {
				return
			}
			err := task()
			wp.ErrorQueue <- err
		case <-wp.done:
			return
		}
	}
}

func (wp *WorkerPool) RunWithTimeout(ctx context.Context) error {
	select {
	case <-ctx.Done():
		wp.Shutdown()
		return ctx.Err()
	case err := <-wp.ErrorQueue:
		if err != nil {
			wp.Shutdown()
			return err
		}
	}
	return nil
}

func (wp *WorkerPool) Submit(task taskFunc) {
	wp.taskQueue <- task
}

func (wp *WorkerPool) Shutdown() {
	close(wp.done)
	close(wp.taskQueue)
}
