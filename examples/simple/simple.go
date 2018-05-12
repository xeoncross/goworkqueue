package main

import (
	"fmt"
	"time"

	"github.com/xeoncross/goworkqueue"
)

/*
 * Simple example of creating a work queue that gracefully handles shutdowns
 * or failure
 */

// A real worker would be parsing a web page or crunching numbers
func workerFunc(job interface{}, workerID int) {

	fmt.Println("worker", workerID, "processing job", job)
	time.Sleep(1 * time.Second)
	fmt.Println("worker", workerID, "saving job", job)

	// switch v := job.(type) {
	// case string:
	// 	fmt.Println("string:", v)
	// case int, int32, int64:
	// 	fmt.Println("int:", v)
	// case float32:
	// 	fmt.Println("float32:", v)
	// default:
	// 	fmt.Println("unknown")
	// }
}

func main() {

	jobQueueSize := 100
	numberOfWorkers := 3

	queue := goworkqueue.NewQueue(jobQueueSize, numberOfWorkers, workerFunc)

	// Pretend we suddenly need to stop the workers.
	// This might be a SIGTERM or perhaps the workerFunc() called queue.Close()
	go func() {
		time.Sleep(1 * time.Second)
		queue.Close()
		fmt.Println("ABORT!")
	}()

	// We can optionally prefill the work queue
	for j := 1; j <= 20; j++ {
		if ok := queue.Add(fmt.Sprintf("Job %d", j)); !ok {
			break
		}
	}

	// Blocks until queue.Close()
	queue.Run()

	// It's easy to check on the status of the queue
	if queue.Closed() {
		// Always true in this case since it's below queue.Run()
	}

	// Optional, callback for emptying the queue *if* anything remains
	queue.Drain(func(job interface{}) {
		fmt.Printf("'%v' wasn't finished\n", job)
	})

}
