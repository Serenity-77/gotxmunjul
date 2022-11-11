package worker


import (
    "sync"
)


var _ IWorkerConsumer = (*Consumer)(nil)

type Consumer struct {
    handler     func(interface{})
    workerNum   int
    stopWg      sync.WaitGroup
}


func NewConsumer(handler func(interface{}), workerNum int) *Consumer {
    if workerNum <= 0 {
        workerNum = 1
    }

    consumer := &Consumer{
        handler: handler,
        workerNum: workerNum,
    }

    return consumer
}


func (c *Consumer) StartConsuming(queue IWorkerQueueConsumer) {
    for i := 0; i < c.workerNum; i++ {
        c.stopWg.Add(1)
        go c.consumingLoop(queue)
    }
}

func (c *Consumer) StopConsuming() {
    c.stopWg.Wait()
}


func (c *Consumer) consumingLoop(queue IWorkerQueueConsumer) {
    defer c.stopWg.Done()

    for {
        select {
        case item, ok := <- queue.Get():
            if !ok {
                return
            }
            c.handler(item)
        }
    }
}
