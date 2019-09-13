package goworkqueue

import (
	"sync"
	"time"
)

// Queue struct
type Queue struct {
	jobs    chan interface{}
	done    chan bool
	workers chan chan int
	once    sync.Once
}

// NewQueue work queue
func NewQueue(size int, workers int, callback func(interface{}, int)) (q *Queue) {

	q = &Queue{}

	q.jobs = make(chan interface{}, size)
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
			case j := <-q.jobs:
				callback(j, id)
			}
		}

		close(done)
	}()

	return done
}

// Run blocks until the queue is closed
func (q *Queue) Run() {

	// Wait for all workers to be halted
	for w := range q.workers {
		<-w
	}

	// TODO?
	// There seems to be a theoretical chance of a race condition by Add()
	// checking q.done before Close() is called and then trying to send on q.jobs
	// *after* Close() has been called. By closing q.jobs here, instead of in
	// Close(), we avoid this(?) because between these two events all the workers
	// have to stop working which is a much greater timespan then the time
	// between checking q.done and sending on q.jobs
	close(q.jobs)
}

// Drain queue of jobs
func (q *Queue) Drain(callback func(interface{})) {
	for j := range q.jobs {
		callback(j)
	}
}

// Close the work queue
func (q *Queue) Close() {
	q.once.Do(func() {
		close(q.done)
	})
}

// Closed reports if this queue is already closed
func (q *Queue) Closed() bool {
	select {
	case <-q.done:
		return true
	default:
		return false
	}
}

// Add jobs to the queue as long as it hasn't be closed
func (q *Queue) Add(job interface{}) bool {
	// Check the queue is open first
	select {
	case <-q.done:
		return false
	default:
		// While the jobs queue send is blocking, we might shutdown the queue
		select {
		case q.jobs <- job:
			return true
		case <-q.done:
			return false
		}
	}
}

// SleepUntilTimeOrChanActivity (whichever comes first)
func SleepUntilTimeOrChanActivity(t time.Duration, c chan interface{}) {
	select {
	case <-time.After(t):
	case <-c:
	}
}
