package worker

import (
    "sync"
    "testing"
    "github.com/stretchr/testify/assert"
)


func TestQueueRunSimple(t *testing.T) {
    q := NewQueue()
    m := 5

    expected := make([]int, m)

    for i := 0; i < 5; i++ {
        q.Put(i + 1)
        expected[i] = i + 1
    }

    assert.Equal(t, 5, q.Pending())

    values := make([]int, m)

    wg := sync.WaitGroup{}
    wg.Add(1)

    go func() {
        defer wg.Done()
        c := 0
        for {
            values[c] = (<- q.Get()).(int)
            c += 1
            if c == m {
                break
            }
        }
    }()

    wg.Wait()

    assert.Equal(t, 0, q.Pending())
    assert.ElementsMatch(t, expected, values)
}

func TestQueueClose(t *testing.T) {
    q := NewQueue()
    closed := 0

    wg := sync.WaitGroup{}

    wg.Add(1)

    go func() {
        defer wg.Done()
        for {
            _, ok := <- q.Get()
            if !ok {
                break
            }
        }
        closed = 1
    }()

    q.Close()
    wg.Wait()

    assert.Equal(t, 1, closed)

    select {
    case _, ok := <- q.closeChan:
        assert.False(t, ok)
    default:
        assert.Fail(t, "Queue Not Closed")
    }
}

func TestQueueCloseFinishPending(t *testing.T) {
    q := NewQueue()
    waitClosed := make(chan struct{})
    wait := make(chan struct{})
    values := []int{}
    closed := 0

    wg := sync.WaitGroup{}
    wg.Add(1)

    go func() {
        defer wg.Done()
        <- q.Get()
        wait <- struct{}{}

        <- waitClosed

        for {
            data, ok := <- q.Get()
            if !ok {
                break
            }
            values = append(values, data.(int))
        }
        closed = 1
    }()

    q.Put(1)
    <- wait
    expected := []int{11, 22, 33, 44, 55}

    for _, v := range expected {
        q.Put(v)
    }

    assert.Equal(t, len(expected), q.Pending())

    waitClosed <- struct{}{}
    q.Close()

    wg.Wait()

    assert.Equal(t, 1, closed)
    assert.ElementsMatch(t, expected, values)

    select {
    case _, ok := <- q.closeChan:
        assert.False(t, ok)
    default:
        assert.Fail(t, "Queue Not Closed")
    }
}

func TestQueueCloseEmpty(t *testing.T) {
    q := NewQueue()
    q.Close()
    select {
    case _, ok := <- q.closeChan:
        assert.False(t, ok)
    default:
        assert.Fail(t, "Queue Not Closed")
    }
    select {
    case _, ok := <- q.waiting:
        assert.False(t, ok)
    default:
        assert.Fail(t, "Queue Not Closed")
    }
}
