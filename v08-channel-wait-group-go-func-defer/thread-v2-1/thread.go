package thread_v2_1

import (
	"log"
	"sync"
	"time"
)

type Pool struct {
	MaxThread int
	workers   chan int
	waitGroup sync.WaitGroup
	tasks     []func()
}

func (tp *Pool) Init() {
	tp.workers = make(chan int, tp.MaxThread)
	tp.waitGroup = sync.WaitGroup{}
}

func (tp *Pool) Fetch() int {
	return <-tp.workers
}

func (tp *Pool) Release(index int) {
	tp.workers <- index
}

func (tp *Pool) Start() {
	for i := 0; i < tp.MaxThread; i++ {
		tp.workers <- i
	}

	for i, fun := range tp.tasks {
		function := fun
		index := i

		tp.waitGroup.Add(1)
		workerId := tp.Fetch()
		go func() {
			defer tp.Release(workerId)
			defer tp.waitGroup.Done()

			log.Printf("Worker [%d] started at %d, task id is [%d]\n", workerId, time.Now().Unix(), index)
			function()
			log.Printf("Worker [%d] finished at %d, task id is [%d]\n", workerId, time.Now().Unix(), index)
		}()
	}
}

func (tp *Pool) Wait() {
	tp.waitGroup.Wait()
	close(tp.workers)
}

func (tp *Pool) AddTask(index int, task func()) {
	tp.tasks = append(tp.tasks, task)
}
