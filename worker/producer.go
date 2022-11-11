package worker

import (
    "time"
    txUtils "github.com/serenity-77/bagudung/utils"
)


var _ IWorkerProducer = (*Producer)(nil)

type IProducerHandler interface {
    Enqueue(chan <- interface{})
    Stop()
}

type Producer struct {
    handler     IProducerHandler
    chanQueue   chan interface{}
    closed      chan struct{}
}


func NewProducer(handler IProducerHandler) *Producer {
    producer := &Producer{
        handler:    handler,
        chanQueue:  make(chan interface{}),
        closed:     make(chan struct{}),
    }
    return producer
}


func (p *Producer) StartProducing(queue IWorkerQueueProducer) {
    defer close(p.closed)

    go p.handler.Enqueue(p.chanQueue)

    for {
        select {
        case item, ok := <- p.chanQueue:
            if !ok {
                return
            }
            queue.Put(item)
        }
    }
}


func (p *Producer) StopProducing() {
    p.handler.Stop()
    close(p.chanQueue)
    <- p.closed
    p.handler = nil
    p.chanQueue = nil
    p.closed = nil
}


var _ IProducerHandler = (*IntervalProducerHandler)(nil)

type IntervalProducerHandler struct {
    fetch       func() []interface{}
    interval    time.Duration
    clock       txUtils.IClock
    enqueueNow  bool
    stop        chan struct{}
    stopWait    chan struct{}
}


func NewIntervalProducerHandler(fetch func() []interface{}, interval time.Duration, enqueueNow bool) *IntervalProducerHandler {
    handler := &IntervalProducerHandler{
        fetch:      fetch,
        interval:   interval,
        clock:      txUtils.NewRealClock(),
        enqueueNow: enqueueNow,
        stop:       make(chan struct{}),
        stopWait:   make(chan struct{}),
    }
    return handler
}

func (self *IntervalProducerHandler) Enqueue(chanQueue chan <- interface{}) {
    if self.enqueueNow {
        self.fetchAndEnqueue(chanQueue)
        self.enqueueNow = false
    }

    self.enqueueLoop(chanQueue)
}

func (self *IntervalProducerHandler) fetchAndEnqueue(chanQueue chan <- interface{}) {
    items := self.fetch()
    for _, item := range items {
        chanQueue <- item
    }
}

func (self *IntervalProducerHandler) enqueueLoop(chanQueue chan <- interface{}) {
    defer close(self.stopWait)

    intervalTimer := self.clock.Timer(self.interval)

    for {
        select {
        case <- intervalTimer.C:
            self.fetchAndEnqueue(chanQueue)
            intervalTimer.Reset(self.interval)
        case <- self.stop:
            intervalTimer.Stop()
            return
        }
    }
}

func (self *IntervalProducerHandler) Stop() {
    close(self.stop)
    <- self.stopWait
    self.fetch = nil
    self.clock = nil
    self.stop = nil
    self.stopWait = nil
}
