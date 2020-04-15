package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {
	maxTask := 10

	maxThread := 3
	ch := make(chan int, maxThread)
	for i := 0; i < maxThread; i++ {
		threadId := i
		go func() {
			wg.Add(1)
			defer wg.Done()

			log.Printf("Worker [%d] started at %d\n", threadId, time.Now().Unix())

			for taskId := range ch {
				seconds := 1 + rand.Intn(9)
				log.Printf("Task [%d] will sleep %d seconds\n", taskId, seconds)
				time.Sleep(time.Second * time.Duration(seconds))
				log.Printf("Task [%d] finished", taskId)
			}

			log.Printf("Worker [%d] finished at %d\n", threadId, time.Now().Unix())
		}()
	}

	for i := 0; i < maxTask; i++ {
		ch <- i
	}

	close(ch)
	wg.Wait()
}
