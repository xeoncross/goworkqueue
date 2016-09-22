# GoWorkQueue

Super simple job queue with managed workers. No locking mutexes, only channel communication. Perfect for jobs like crawling websites.

## Install

    go get github.com/xeoncross/goworkqueue

## Usage

Create a new queue instance with a callback for each job you want run.

      queue := goworkqueue.Queue{}
      queue.Init(1000, 5, func(job string, workerId int) {
        fmt.Println(job, workerId)
      })
      queue.Run() // Blocks until queue.Close() is called

See the example/example.go for more information.

Released Free under the MIT license http://davidpennington.me
