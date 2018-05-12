# GoWorkQueue

Super simple single-node job queue with managed workers. Perfect for small jobs like digesting streams or simple crawl jobs.

## Install

    go get github.com/xeoncross/goworkqueue

## Usage

Create a new queue instance with a callback for each job you want run.

    queue := goworkqueue.NewQueue(1000, 5, func(job interface{}, workerID int) {
      fmt.Println("Processing", job)
    })
    queue.Add("one")  // anything can add "jobs" to process
    queue.Run()       // Blocks until queue.Close() is called

See the example/example.go for more information.

Released Free under the MIT license http://davidpennington.me
