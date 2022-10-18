package jobs

import (
    // "fmt"
)

type Queue struct {
    queue       chan interface{}
    pending     []interface{}
    waiting     chan interface{}
    lchan       chan int
    closeChan   chan struct{}
}

func NewQueue() *Queue {
    q := &Queue{}
    q.queue = make(chan interface{})
    q.waiting = make(chan interface{})
    q.lchan = make(chan int)
    q.closeChan = make(chan struct{})
    go q._workerLoop()
    return q
}


func (q *Queue) _workerLoop() {
    // make sure workerLoop goroutine is stopped
    // so that no race can occurs.
    defer close(q.closeChan)
    defer close(q.waiting)

Loop:
    for {
        if len(q.pending) > 0 {
            if !q._processPending() {
                break
            }
        } else {
            select {
            case q.lchan <- len(q.pending):
            case data, ok := <- q.queue:
                if !ok {
                    break Loop
                }
                select {
                case q.waiting <- data:
                default:
                    q.pending = append(q.pending, data)
                }
            }
        }
    }

    q._finishPending()
}

func (q *Queue) Put(data interface{}) {
    q.queue <- data
}

func (q *Queue) Get() <- chan interface{} {
    return q.waiting
}

func (q *Queue) Pending() int {
    select {
    case <- q.closeChan:
        return len(q.pending)
    case p := <- q.lchan:
        return p
    }
}

func (q *Queue) Close() {
    close(q.queue)
    <- q.closeChan
}

func (q *Queue) _processPending() bool {
    select {
    case q.lchan <- len(q.pending):
    case data, ok := <- q.queue:
        if !ok {
            return false
        }
        q.pending = append(q.pending, data)
    case q.waiting <- q.pending[0]:
        q.pending = q.pending[1:]
    }
    return true
}

func (q *Queue) _finishPending() {
    for len(q.pending) > 0 {
        select {
        case q.waiting <- q.pending[0]:
            q.pending = q.pending[1:]
        case q.lchan <- len(q.pending):
        }
    }
}
