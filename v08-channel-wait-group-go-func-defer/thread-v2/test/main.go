package main

import (
	"../../thread-v2"
	"log"
	"math/rand"
	"time"
)

func main() {
	tp := Thread.Pool{MaxThread: 3}
	tp.Init()

	tasksCount := 10
	for i := 0; i < tasksCount; i++ {
		tmpI := i
		tp.AddTask(tmpI, func() {
			seconds := rand.Intn(9) + 1
			log.Printf("Task [%d] will sleep %d seconds", tmpI, seconds)
			time.Sleep(time.Second * time.Duration(seconds))
			log.Printf("Task [%d] finished", tmpI)
		})
	}

	tp.Start()
	tp.Wait()
}
