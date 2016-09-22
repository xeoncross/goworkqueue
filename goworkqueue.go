package goworkqueue

type Queue struct {
	Jobs    chan string
	done    chan bool
	workers chan chan int
}

func (q *Queue) worker(id int, callback func(string, int)) (done chan int) {
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

func (q *Queue) Init(size int, workers int, callback func(string, int)) {

	q.Jobs = make(chan string, size)
	q.done = make(chan bool)
	q.workers = make(chan chan int, workers)

	for w := 1; w <= workers; w++ {
		q.workers <- q.worker(w, callback)
	}

	close(q.workers)
}

// Run blocks until the "done" channel is closed on the queue
func (q *Queue) Run() {

	// Wait for workers to be halted
	for w := range q.workers {
		<-w
	}

	// Nothing should still be mindlessly adding jobs
	close(q.Jobs)

}

// Allow the queueue to be drained after it is closed
func (q *Queue) Drain(callback func(string)) {
	for j := range q.Jobs {
		callback(j)
	}
}

func (q *Queue) Close() {
	close(q.done)
}
