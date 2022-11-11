package worker


import (
    "testing"
    "github.com/stretchr/testify/assert"
)


type dummyProducer struct {
    started chan struct{}
    stopped bool
}

func (dp *dummyProducer) StartProducing(queue IWorkerQueueProducer) {
    dp.started <- struct{}{}
}

func (dp *dummyProducer) StopProducing() {
    dp.stopped = true
}

type dummyConsumer struct {
    started chan struct{}
    stopped bool
}

func (dc *dummyConsumer) StartConsuming(queue IWorkerQueueConsumer) {
    dc.started <- struct{}{}
}

func (dc *dummyConsumer) StopConsuming() {
    dc.stopped = true
}

func assertWorkerStarted(t *testing.T, w *Worker) {
    assert.NotNil(t, w.queue)
    assert.NotNil(t, w.producer)
    assert.NotNil(t, w.consumer)
}

func assertWorkerStopped(t *testing.T, w *Worker, q *Queue) {
    assert.Nil(t, w.queue)
    assert.Nil(t, w.producer)
    assert.Nil(t, w.consumer)
    <- q.closeChan
    <- q.waiting
}

func TestWorkerBasic(t *testing.T) {
    producer := &dummyProducer{started: make(chan struct{})}
    consumer := &dummyConsumer{started: make(chan struct{})}

    worker := NewWorker(producer, consumer)

    assert.NotNil(t, worker)

    assertWorkerStarted(t, worker)

    <- producer.started
    <- consumer.started
}

type dummyProducer1 struct {}

func (dp *dummyProducer1) StartProducing(queue IWorkerQueueProducer) {
    values := []int{11, 22, 33, 44, 55}
    go func() {
        for _, value := range values {
            queue.Put(value)
        }
    }()
}

func (dp *dummyProducer1) StopProducing(){}


type dummyConsumer1 struct {
    wait    chan struct{}
    data    []int
}

func (dc *dummyConsumer1) StartConsuming(queue IWorkerQueueConsumer) {
    go func() {
        for data := range queue.Get() {
            dc.data = append(dc.data, data.(int))
            if len(dc.data) == 5 {
                dc.wait <- struct{}{}
                return
            }
        }
    }()
}

func (dc *dummyConsumer1) StopConsuming(){}

func TestWorkerSimpleProduceConsume(t *testing.T) {
    producer := &dummyProducer1{}
    consumer := &dummyConsumer1{wait: make(chan struct{})}

    worker := NewWorker(producer, consumer)

    assert.NotNil(t, worker)

    assertWorkerStarted(t, worker)

    <- consumer.wait

    assert.ElementsMatch(t, []int{11, 22, 33, 44, 55}, consumer.data)
}


func TestWorkerStop(t *testing.T) {
    producer := &dummyProducer{started: make(chan struct{})}
    consumer := &dummyConsumer{started: make(chan struct{})}

    worker := NewWorker(producer, consumer)

    <- producer.started
    <- consumer.started

    queue := worker.queue

    assert.NotNil(t, worker)

    assertWorkerStarted(t, worker)

    worker.Stop()

    assertWorkerStopped(t, worker, queue)

    assert.True(t, producer.stopped)
    assert.True(t, consumer.stopped)
}
