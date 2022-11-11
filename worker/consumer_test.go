package worker

import (
    "sync"
    "testing"
    "github.com/stretchr/testify/assert"
)

func assertConsumer(t *testing.T, c *Consumer, expectedWorkerNum int) {
    assert.NotNil(t, c)
    assert.NotNil(t, c.handler)
    assert.Equal(t, expectedWorkerNum, c.workerNum)
}


func TestConsumer(t *testing.T) {
    consumer := NewConsumer(func(data interface{}){}, 3)
    assertConsumer(t, consumer, 3)
}

type testData1 struct {
    value   int
    wait    chan int
}

func TestConsumerStartConsuming(t *testing.T) {
    wg := sync.WaitGroup{}
    queue := NewQueue()

    consumer := NewConsumer(func(data interface{}){
        if data == nil {
            return
        }
        x := data.(testData1)
        wg.Done()
        x.wait <- x.value
    }, 3)

    assertConsumer(t, consumer, 3)

    consumer.StartConsuming(queue)

    items := []testData1{
        {1, make(chan int)},
        {2, make(chan int)},
        {3, make(chan int)},
    }

    wg.Add(len(items))

    for _, item := range items {
        queue.Put(item)
    }

    wg.Wait()

    // Assert to check that all workers goroutines are busy.
    for i := 0; i < 5; i++ {
        queue.Put(nil)
    }
    assert.Equal(t, 5, queue.Pending())

    values := []int{}
    for _, item := range items {
        value := <- item.wait
        values = append(values, value)
    }

    assert.ElementsMatch(t, []int{1, 2, 3}, values)
}

type testData2 struct {
    wait    chan struct{}
}

func TestConsumerStopConsuming(t *testing.T) {
    wg := sync.WaitGroup{}
    queue := NewQueue()

    consumer := NewConsumer(func(data interface{}){
        x := data.(testData2)
        wg.Done()
        <- x.wait
    }, 3)

    assertConsumer(t, consumer, 3)

    consumer.StartConsuming(queue)

    items := []testData2{
        {make(chan struct{})},
        {make(chan struct{})},
        {make(chan struct{})},
    }

    wg.Add(len(items))

    for _, item := range items {
        queue.Put(item)
    }

    wg.Wait()

    queue.Close()
    <- queue.closeChan
    <- queue.waiting

    go func() {
        for _, item := range items {
            item.wait <- struct{}{}
        }
    }()

    consumer.StopConsuming()
}
