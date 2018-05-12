package goworkqueue

import (
	"sync"
	"time"
)

// Queue struct
type Queue struct {
	Jobs    chan interface{}
	done    chan bool
	workers chan chan int
	mux     sync.Mutex
}

// NewQueue work queue
func NewQueue(size int, workers int, callback func(interface{}, int)) (q *Queue) {

	q = &Queue{}

	q.Jobs = make(chan interface{}, size)
	q.done = make(chan bool)
	q.workers = make(chan chan int, workers)

	for w := 1; w <= workers; w++ {
		q.workers <- q.worker(w, callback)
	}

	close(q.workers)
	return
}

func (q *Queue) worker(id int, callback func(interface{}, int)) (done chan int) {
	done = make(chan int)

	go func() {
	work:
		for {
			select {
			case <-q.done:
				break work
			case j := <-q.Jobs:
				callback(j, id)
			}
		}

		done <- id
		close(done)
	}()

	return done
}

// Run blocks until the queue is closed
func (q *Queue) Run() {

	// Wait for workers to be halted
	for w := range q.workers {
		<-w
	}

	// Nothing should still be mindlessly adding jobs
	close(q.Jobs)
}

// Drain queue of jobs
func (q *Queue) Drain(callback func(interface{})) {
	for j := range q.Jobs {
		callback(j)
	}
}

// Close the work queue
func (q *Queue) Close() {
	q.mux.Lock()
	close(q.done)
	q.mux.Unlock()
}

// Closed reports if this queue is already closed
func (q *Queue) Closed() bool {
	q.mux.Lock()
	defer q.mux.Unlock()

	select {
	case <-q.done:
		return true
	default:
		return false
	}
}

// Add jobs to the queue as long as it hasn't be closed
func (q *Queue) Add(job interface{}) (ok bool) {
	q.mux.Lock()
	select {
	case <-q.done:
		ok = false
	case q.Jobs <- job:
		ok = true
	}
	q.mux.Unlock()
	return
}

// SleepUntilTimeOrChanActivity (whichever comes first)
func SleepUntilTimeOrChanActivity(t time.Duration, c chan interface{}) {
	select {
	case <-time.After(t):
	case <-c:
	}
}
