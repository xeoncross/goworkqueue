package goworkqueue

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// func TestCloseSend(t *testing.T) {
//
// 	c := make(chan struct{})
//
// 	go func() {
// 		time.Sleep(time.Microsecond)
// 		close(c)
// 	}()
//
// 	// What happens?
// 	c <- struct{}{}
//
// }

func TestQueue(t *testing.T) {

	// 1000 job queue with 100 workers
	workers := 100

	queue := NewQueue(1000, workers, func(job interface{}, workerID int) {
		// time.Sleep(time.Millisecond)
	})

	// Pretend we suddenly need to stop the workers.
	// This might be a SIGTERM or perhaps the workerFunc() called queue.Close()
	go func() {
		time.Sleep(5 * time.Millisecond)
		queue.Close()
	}()

	var jobs int64

	var group sync.WaitGroup

	group.Add(workers)

	for j := 0; j < workers; j++ {

		go func(id int) {
			var i int
			for {
				i++
				atomic.AddInt64(&jobs, 1)
				if ok := queue.Add(i); !ok {
					break
				}
			}

			group.Done()
			// fmt.Printf("%d: %d jobs\n", id, i)
		}(j)
	}

	// Blocks until queue.Close()
	queue.Run()

	// ensure all goroutines are finished
	group.Wait()

	// fmt.Printf("%d jobs processed\n", atomic.LoadInt64(&jobs))

}
