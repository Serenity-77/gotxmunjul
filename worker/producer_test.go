package worker

import (
    "time"
    "testing"
    "github.com/stretchr/testify/assert"
    txUtils "github.com/serenity-77/bagudung/utils"
)

func assertProducer(t *testing.T, p *Producer) {
    assert.NotNil(t, p)
    assert.NotNil(t, p.handler)
    assert.NotNil(t, p.chanQueue)
    assert.NotNil(t, p.closed)
}

func assertProducerStopped(t *testing.T, p *Producer) {
    assert.Nil(t, p.handler)
    assert.Nil(t, p.chanQueue)
    assert.Nil(t, p.closed)
}

type dummyHandler1 struct {

}

func (dh1 *dummyHandler1) Enqueue(queue chan <- interface{}) {}
func (dh1 *dummyHandler1) Stop() {}


func TestProducer(t *testing.T) {
    producer := NewProducer(&dummyHandler1{})
    assertProducer(t, producer)
}


type dummyHandler2 struct {
    done    chan struct{}
}

func (dh2 *dummyHandler2) Enqueue(chanQueue chan <- interface{}) {
    defer close(dh2.done)
    for i := 0; i < 5; i++ {
        chanQueue <- struct{}{}
    }
}

func (dh2 *dummyHandler2) Stop() {

}

func TestProducerStartProducing(t *testing.T) {
    queue := NewQueue()
    handler := &dummyHandler2{make(chan struct{})}
    producer := NewProducer(handler)

    assertProducer(t, producer)

    go producer.StartProducing(queue)

    <- handler.done
    close(producer.chanQueue)
    <- producer.closed

    assert.Equal(t, 5, queue.Pending())
}

type dummyHandler3 struct {
    running chan struct{}
    stopped chan struct{}
}

func (dh3 *dummyHandler3) Enqueue(chanQueue chan <- interface{}) {
    close(dh3.running)
}

func (dh3 *dummyHandler3) Stop() {
    close(dh3.stopped)
}

func TestProducerStopProducing(t *testing.T) {
    queue := NewQueue()
    handler := &dummyHandler3{make(chan struct{}), make(chan struct{})}
    producer := NewProducer(handler)

    assertProducer(t, producer)

    go producer.StartProducing(queue)

    <- handler.running

    producer.StopProducing()

    <- handler.stopped

    assertProducerStopped(t, producer)
}


func assertIntervalProducerHandler(t *testing.T, h *IntervalProducerHandler, enqueueNow bool, expectedInterval time.Duration) {
    assert.NotNil(t, h)
    assert.NotNil(t, h.fetch)
    assert.NotNil(t, h.clock)
    assert.Equal(t, expectedInterval, h.interval)
    if !enqueueNow {
        assert.False(t, h.enqueueNow)
    } else {
        assert.True(t, h.enqueueNow)
    }
}

func assertIntervalProducerHandlerStop(t *testing.T, h *IntervalProducerHandler) {
    assert.Nil(t, h.fetch)
    assert.Nil(t, h.clock)
    assert.Nil(t, h.stop)
    assert.Nil(t, h.stopWait)
}

func TestIntervalProducerHandler(t *testing.T) {
    handler := NewIntervalProducerHandler(func() []interface{}{
        return nil
    }, 2 * time.Second, false)
    assertIntervalProducerHandler(t, handler, false, 2 * time.Second)
}

func TestIntervalProducerHandlerEnqueueNow(t *testing.T) {
    fetchItems := [][]int{
        []int{11, 12, 13, 14, 15},
        []int{},
        []int{21, 22, 23, 24, 25, 26, 27},
        nil,
    }

    fetch := func() []interface{} {
        items := fetchItems[0]
        fetchItems = fetchItems[1:]

        results := []interface{}{}
        for len(items) > 0 {
            results = append(results, items[0])
            items = items[1:]
        }
        return results
    }

    chanQueue := make(chan interface{}, 100)

    drainQueue := func() {
        for len(chanQueue) > 0 {
            <- chanQueue
        }
    }

    handler := NewIntervalProducerHandler(fetch, 2 * time.Second, true)

    assertIntervalProducerHandler(t, handler, true, 2 * time.Second)

    clock := txUtils.NewFakeClock()

    handler.clock = clock

    go handler.Enqueue(chanQueue)

    clock.WaitUntilBlock(1)

    assert.Equal(t, 3, len(fetchItems))

    assert.Equal(t, 5, len(chanQueue))
    drainQueue()
    assert.Equal(t, 0, len(chanQueue))

    assert.False(t, handler.enqueueNow)

    intervalTimer := clock.GetTimer(0)

    assert.Equal(t, clock.RightNow() + int64(2 * time.Second), intervalTimer.ExpireAt())

    for len(fetchItems) > 0 {
        l := len(fetchItems[0])
        clock.Advance(2 * time.Second)
        clock.WaitUntilBlock(1)
        assert.Equal(t, l, len(chanQueue))
        drainQueue()
        assert.Equal(t, clock.RightNow() + int64(2 * time.Second), intervalTimer.ExpireAt())
    }
}

func TestIntervalProducerHandlerEnqueue(t *testing.T) {
    fetch := func() []interface{} {
        return []interface{}{1, 2, 3}
    }

    chanQueue := make(chan interface{}, 3)

    handler := NewIntervalProducerHandler(fetch, 2 * time.Second, false)
    assertIntervalProducerHandler(t, handler, false, 2 * time.Second)

    clock := txUtils.NewFakeClock()

    handler.clock = clock

    go handler.Enqueue(chanQueue)

    clock.WaitUntilBlock(1)

    intervalTimer := clock.GetTimer(0)

    assert.Equal(t, clock.RightNow() + int64(2 * time.Second), intervalTimer.ExpireAt())

    clock.Advance(2 * time.Second)
    clock.WaitUntilBlock(1)

    assert.Equal(t, 3, len(chanQueue))

    assert.Equal(t, clock.RightNow() + int64(2 * time.Second), intervalTimer.ExpireAt())
}

func TestIntervalProducerHandlerStop(t *testing.T) {
    chanQueue := make(chan interface{}, 3)

    handler := NewIntervalProducerHandler(func() []interface{}{return nil}, 3 * time.Second, false)
    assertIntervalProducerHandler(t, handler, false, 3 * time.Second)

    clock := txUtils.NewFakeClock()

    handler.clock = clock

    go handler.Enqueue(chanQueue)

    clock.WaitUntilBlock(1)

    intervalTimer := clock.GetTimer(0)

    assert.Equal(t, clock.RightNow() + int64(3 * time.Second), intervalTimer.ExpireAt())

    handler.Stop()

    assert.True(t, intervalTimer.Stopped())
    assertIntervalProducerHandlerStop(t, handler)
}

func TestIntervalProducerHandlerStopWhileFetch(t *testing.T) {
    chanQueue := make(chan interface{}, 3)
    fetchReady := make(chan struct{})
    fetchDo := make(chan struct{})

    fetch := func() []interface{} {
        close(fetchReady)
        <- fetchDo
        return nil
    }

    handler := NewIntervalProducerHandler(fetch, 3 * time.Second, false)

    assertIntervalProducerHandler(t, handler, false, 3 * time.Second)

    clock := txUtils.NewFakeClock()

    handler.clock = clock

    go handler.Enqueue(chanQueue)

    clock.WaitUntilBlock(1)

    intervalTimer := clock.GetTimer(0)

    assert.Equal(t, clock.RightNow() + int64(3 * time.Second), intervalTimer.ExpireAt())

    clock.Advance(3 * time.Second)
    <- fetchReady

    go func() {
        close(fetchDo)
    }()

    handler.Stop()

    assert.True(t, intervalTimer.Stopped())
    assertIntervalProducerHandlerStop(t, handler)
}
