package main

import (
	"../../thread-v2-fix"
	"log"
	"math/rand"
	"time"
)

func main() {
	tp := Thread.Pool{MaxThread: 3}
	tp.Prepare(func(param interface{}) {
		taskId := param.(int)

		seconds := rand.Intn(9) + 1
		log.Printf("Task [%d] will sleep %d seconds", taskId, seconds)

		time.Sleep(time.Second * time.Duration(seconds))

		log.Printf("Task [%d] finished", taskId)
	})

	tasksCount := 10
	for i := 0; i < tasksCount; i++ {
		tp.RunWith(i)
	}

	tp.Wait()
}
