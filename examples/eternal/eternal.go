package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xeoncross/goworkqueue"
)

/*
 * Work queue that runs forever generating and processing work
 */

func main() {

	workerFunc := func(job interface{}, workerId int) {
		if id, ok := job.(int64); ok {
			fmt.Printf("Processing ID %d\n", id)
			time.Sleep(time.Millisecond * 500)
		}
	}

	jobQueueSize := 10
	numberOfWorkers := 3

	queue := goworkqueue.NewQueue(jobQueueSize, numberOfWorkers, workerFunc)

	// Abort when we press CTRL+C (go run...) or send a kill -9 (go build...)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for _ = range c {
			queue.Close()
			log.Println("Interrupt / SIGTERM received. Stopping...")
		}
	}()

	// Forever, we add work to the queue to be processed.
	// If queue.Jobs is full, this will halt until the workers
	// make more room in the queue - so our backlog is under control.
	go func() {
		var id int64
		for {
			// Here you could fetch data from a queue or database
			id++

			// If we can't add a job, the queue must be closed/closing
			if ok := queue.Add(id); !ok {
				return
			}
		}

	}()

	// Blocks until queue.Close()
	queue.Run()

	// Optional, callback for emptying the queue *if* anything remains
	queue.Drain(func(job interface{}) {
		fmt.Printf("'%v' wasn't finished\n", job)
	})

}
