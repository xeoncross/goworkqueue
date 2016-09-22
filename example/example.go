package main

import (
  "fmt"
  "time"
  "github.com/xeoncross/goworkqueue"
)

// A real worker would be parsing a web page or crunching numbers
func workerFunc(job string, workerId int) {
  fmt.Println("worker", workerId, "processing job", job)
  time.Sleep(1 * time.Second)
  fmt.Println("worker", workerId, "saving job", job)
}


func main() {

    jobQueueSize := 100
    numberOfWorkers := 3

    queue := goworkqueue.Queue{}
    queue.Init(jobQueueSize, numberOfWorkers, workerFunc)

    // Pretend we suddenly need to stop the workers.
    // This might be a SIGTERM or perhaps the workerFunc() called queue.Close()
    go func() {
        time.Sleep(1 * time.Second)
        queue.Close()
        fmt.Println("ABORT!")
    }()

    // We can optionally prefill the work queue
    for j := 1; j <= 20; j++ {
      queue.Jobs <- fmt.Sprintf("Job %d", j)
    }

    // Blocks until queue.Close()
    queue.Run()

    // Optional, callback for emptying the queue *if* anything remains
    queue.Drain(func(job string) {
      fmt.Printf("'%s' wasn't finished\n", job)
    })

}
