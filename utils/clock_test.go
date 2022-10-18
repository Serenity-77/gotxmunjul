package utils

import (
    "time"
    "sync"
    "testing"
    "github.com/stretchr/testify/assert"
)


func testCall1(fc *FakeClock, wg *sync.WaitGroup) {
    go func() {
        t := fc.Timer(2 * time.Second)
        <- t.C
        wg.Done()
    }()
}

func testCall2(fc *FakeClock, wg *sync.WaitGroup) {
    go func() {
        t := fc.Timer(2 * time.Second)
        <- t.C
        t.Reset(2 * time.Second)
        <- t.C
        t.Reset(1 * time.Second)
        <- t.C
        t.Reset(3 * time.Second)
        <- t.C
        t.Reset(2 * time.Second)
        <- t.C
        t.Reset(4 * time.Second)
        <- t.C
        wg.Done()
    }()
}


func TestFakeClockTimer(t *testing.T) {
    fc := NewFakeClock()
    wg := sync.WaitGroup{}

    assert.Equal(t, 0, len(fc.timers))

    wg.Add(3)
    testCall1(fc, &wg)
    testCall1(fc, &wg)
    testCall1(fc, &wg)

    rightNow := fc.RightNow()

    fc.WaitUntilBlock(3)

    assert.Equal(t, fc.timers[0].expireAt, rightNow + int64(2 * time.Second))
    assert.Equal(t, fc.timers[1].expireAt, rightNow + int64(2 * time.Second))
    assert.Equal(t, fc.timers[2].expireAt, rightNow + int64(2 * time.Second))
    assert.Equal(t, 3, len(fc.timers))

    fc.Advance(2 * time.Second)

    assert.Equal(t, 0, len(fc.timers))
    assert.Equal(t, fc.RightNow(), rightNow + int64(2 * time.Second))

    wg.Wait()
}

func TestFakeClockAdvance(t *testing.T) {
    fc := NewFakeClock()
    wg := sync.WaitGroup{}

    rightNow := fc.RightNow()

    wg.Add(1)
    testCall2(fc, &wg)

    fc.WaitUntilBlock(1)

    assert.Equal(t, 1, len(fc.timers))
    assert.Equal(t, fc.timers[0].expireAt, rightNow + int64(2 * time.Second))

    fc.Advance(1500 * time.Millisecond)

    assert.Equal(t, fc.RightNow(), rightNow + int64(1500 * time.Millisecond))
    assert.Equal(t, 1, len(fc.timers))
    assert.Equal(t, fc.timers[0].expireAt, rightNow + int64(2 * time.Second))

    rightNow = fc.RightNow()
    fc.Advance(500 * time.Millisecond)

    assert.Equal(t, fc.RightNow(), rightNow + int64(500 * time.Millisecond))

    fc.WaitUntilBlock(1)
    assert.Equal(t, 1, len(fc.timers))

    rightNow = fc.RightNow()
    assert.Equal(t, fc.timers[0].expireAt, rightNow + int64(2 * time.Second))

    fc.Advance(2 * time.Second)

    assert.Equal(t, fc.RightNow(), rightNow + int64(2 * time.Second))
    fc.WaitUntilBlock(1)
    assert.Equal(t, 1, len(fc.timers))

    rightNow = fc.RightNow()
    assert.Equal(t, fc.timers[0].expireAt, rightNow + int64(time.Second))

    x := []struct{
        d time.Duration
    }{
        {1 * time.Second},
        {3 * time.Second},
        {2 * time.Second},
        {4 * time.Second},
    }

    for len(x) > 0 {
        y := x[0]
        x = x[1:]

        assert.Equal(t, fc.timers[0].expireAt, rightNow + int64(y.d))
        fc.Advance(y.d)
        assert.Equal(t, fc.RightNow(), rightNow + int64(y.d))

        if len(x) > 0 {
            fc.WaitUntilBlock(1)
            assert.Equal(t, 1, len(fc.timers))
            rightNow = fc.RightNow()
        }
    }

    wg.Wait()
}
