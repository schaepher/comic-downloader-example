package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {
	maxThread := 3
	ch := make(chan int, maxThread)

	taskCount := 10
	for i := 0; i < taskCount; i++ {
		tmpId := i
		go func(taskId int) {
			wg.Add(1)
			defer wg.Done()

			workerId := <-ch
			log.Printf("Worker [%d] started at %d, task id is [%d]\n", workerId, time.Now().Unix(), taskId)

			seconds := 1 + rand.Intn(9)
			log.Printf("Task [%d] will sleep %d seconds\n", taskId, seconds)
			time.Sleep(time.Second * time.Duration(seconds))
			log.Printf("Task [%d] finished", taskId)

			log.Printf("Worker [%d] finished at %d\n", workerId, time.Now().Unix())

			ch <- workerId
		}(tmpId)
	}

	for i := 1; i <= maxThread; i++ {
		ch <- i
	}

	wg.Wait()
	close(ch)
}
