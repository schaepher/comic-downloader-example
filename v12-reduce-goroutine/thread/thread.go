package Thread

import (
	"log"
	"sync"
	"time"
)

type Pool struct {
	MaxThread int
	chParams  chan interface{}
	waitGroup sync.WaitGroup
	function  func(param interface{})
}

func (tp *Pool) Prepare(function func(item interface{})) {
	tp.chParams = make(chan interface{}, tp.MaxThread)
	tp.waitGroup = sync.WaitGroup{}
	tp.function = function

	for i := 0; i < tp.MaxThread; i++ {
		workerId := i
		go func() {
			tp.waitGroup.Add(1)
			defer tp.waitGroup.Done()

			log.Printf("Worker [%d] started at %d\n", workerId, time.Now().Unix())
			for param := range tp.chParams {
				tp.function(param)
			}
			log.Printf("Worker [%d] finished at %d\n", workerId, time.Now().Unix())
		}()
	}
}

func (tp *Pool) RunWith(param interface{}) {
	tp.chParams <- param
}

func (tp *Pool) Wait() {
	close(tp.chParams)
	tp.waitGroup.Wait()
}
