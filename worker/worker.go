package worker


import (
    "github.com/sirupsen/logrus"
)

type IWorkerProducer interface {
    StartProducing(IWorkerQueueProducer)
    StopProducing()
}

type IWorkerConsumer interface {
    StartConsuming(IWorkerQueueConsumer)
    StopConsuming()
}


type IWorkerQueueProducer interface {
    Put(interface{})
}

type IWorkerQueueConsumer interface {
    Get()   <- chan interface{}
}

type Worker struct {
    queue       *Queue
    producer    IWorkerProducer
    consumer    IWorkerConsumer
    logger      *logrus.Logger
}

func NewWorker(producer IWorkerProducer, consumer IWorkerConsumer) *Worker {
    worker := &Worker{
        queue:      NewQueue(),
        producer:   producer,
        consumer:   consumer,
    }

    go worker.startConsumer()
    go worker.startProducer()

    return worker
}

func (w *Worker) Stop() {
    w.producer.StopProducing()
    w.queue.Close()
    w.consumer.StopConsuming()
    w.producer = nil
    w.consumer = nil
    w.queue = nil
}

func (w *Worker) startConsumer() {
    w.consumer.StartConsuming(w.queue)
}

func (w *Worker) startProducer() {
    w.producer.StartProducing(w.queue)
}
