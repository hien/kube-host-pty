package util

import (
	"sync"
)

var (
	Workers = &workerGroup{}
)

type Worker func(sigContinue func()) (result interface{}, err error)

type workerGroup struct {
	wg sync.WaitGroup
}

type WorkerChan struct {
	Continue chan struct{}
	Result   chan interface{}
	Error    chan error
}

func (w *WorkerChan) close() {
	close(w.Error)
	close(w.Result)
	close(w.Continue)
}

func newWorkerChan() *WorkerChan {
	return &WorkerChan{
		Continue: make(chan struct{}, 1),
		Result:   make(chan interface{}, 1),
		Error:    make(chan error, 1),
	}
}

func (w *workerGroup) Add(workers ...Worker) []*WorkerChan {
	chList := make([]*WorkerChan, len(workers))

	for i, worker := range workers {
		ch := newWorkerChan()
		chList[i] = ch

		w.wg.Add(1)
		go func(ch *WorkerChan, worker Worker) {
			defer func() {
				ch.close()
				w.wg.Done()
			}()

			if ret, err := worker(func() { ch.Continue <- struct{}{} }); err != nil {
				ch.Error <- err
			} else {
				ch.Result <- ret
			}
		}(ch, worker)
	}
	return chList
}

func (w *workerGroup) wait() {
	w.wg.Wait()
}
