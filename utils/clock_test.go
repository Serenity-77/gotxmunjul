package utils

import (
    "time"
    "sync"
    "testing"
    "github.com/stretchr/testify/assert"
)


func testCall1(fc *FakeClock, wg *sync.WaitGroup, d time.Duration) {
    go func() {
        t := fc.Timer(d)
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
        t.Reset(2 * time.Second)
        <- t.C
        wg.Done()
    }()
}


func TestFakeClockTimer(t *testing.T) {
    fc := NewFakeClock()
    wg := sync.WaitGroup{}

    assert.Equal(t, 0, len(fc.timers))

    wg.Add(3)
    testCall1(fc, &wg, 2 * time.Second)
    testCall1(fc, &wg, 3 * time.Second)
    testCall1(fc, &wg, 4 * time.Second)

    rightNow := fc.RightNow()

    fc.WaitUntilBlock(3)

    assert.Equal(t, fc.GetTimer(0).expireAt, rightNow + int64(2 * time.Second))
    assert.Equal(t, fc.GetTimer(1).expireAt, rightNow + int64(3 * time.Second))
    assert.Equal(t, fc.GetTimer(2).expireAt, rightNow + int64(4 * time.Second))

    assert.Equal(t, 3, len(fc.timers))

    durations := []time.Duration{
        2 * time.Second,
        3 * time.Second,
        4 * time.Second,
    }

    duration := durations[0]
    expectedLen := len(durations) - 1

    for i := 0; i < len(durations); i++ {
        fc.Advance(duration)
        assert.Equal(t, expectedLen, len(fc.timers))
        assert.Equal(t, fc.RightNow(), rightNow + int64(duration))
        duration = 1 * time.Second
        expectedLen--
        rightNow = fc.RightNow()
    }

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
        {2 * time.Second},
    }

    for len(x) > 0 {
        y := x[0]
        x = x[1:]

        assert.Equal(t, rightNow + int64(y.d), fc.timers[0].expireAt)
        fc.Advance(y.d)
        assert.Equal(t, rightNow + int64(y.d), fc.RightNow())

        if len(x) > 0 {
            fc.WaitUntilBlock(1)
            assert.Equal(t, 1, len(fc.timers))
            rightNow = fc.RightNow()
        }
    }

    wg.Wait()
}

func testCall3(test *testing.T, fc *FakeClock, wg *sync.WaitGroup, ts *testStop) {
    go func() {
        t := fc.Timer(2 * time.Second)
        ts.timer <- t
        r := t.Stop()
        ts.isStopped <- r
        <- t.C
        ts.advanced = true
        ts.wait <- struct{}{}
        wg.Done()
    }()
}

type testStop struct {
    timer       chan *Timer
    isStopped   chan bool
    wait        chan struct{}
    advanced    bool
}

func TestFakeTimerStop(t *testing.T) {
    fc := NewFakeClock()
    wg := sync.WaitGroup{}
    ts := testStop{
        timer:      make(chan *Timer),
        isStopped:  make(chan bool),
        wait:       make(chan struct{}),
    }

    wg.Add(1)
    testCall3(t, fc, &wg, &ts)

    fc.WaitUntilBlock(1)
    timer := <- ts.timer

    rightNow := fc.RightNow()
    ft := fc.GetTimer(0)

    assert.Equal(t, rightNow + int64(2 * time.Second), ft.ExpireAt())

    r := <- ts.isStopped

    assert.True(t, r)
    assert.True(t, ft.Stopped())

    fc.Advance(2 * time.Second)

    assert.Equal(t, rightNow + int64(2 * time.Second), fc.RightNow())
    assert.False(t, ts.advanced)

    rightNow = fc.RightNow()

    timer.Reset(4 * time.Second)
    fc.WaitUntilBlock(1)

    assert.Equal(t, rightNow + int64(4 * time.Second), ft.ExpireAt())
    assert.False(t, ft.Stopped())

    fc.Advance(4 * time.Second)
    <- ts.wait
    assert.True(t, ts.advanced)
    assert.Equal(t, rightNow + int64(4 * time.Second), fc.RightNow())

    wg.Wait()
}

func testCall4(fc *FakeClock, wg *sync.WaitGroup, tn *testNotifyStopped) {
    go func() {
        defer wg.Done()
        timer := fc.Timer(2 * time.Second)
        <- timer.C
        tn.foo++
        timer.Stop()
    }()
}


type testNotifyStopped struct {
    foo int
}

func TestFakeTimerWaitStop(t *testing.T) {
    fc := NewFakeClock()
    wg := sync.WaitGroup{}
    tn := testNotifyStopped{}

    wg.Add(1)
    testCall4(fc, &wg, &tn)

    fc.WaitUntilBlock(1)

    rightNow := fc.RightNow()
    timer := fc.GetTimer(0)
    stop := timer.WaitStop()

    assert.NotEmpty(t, timer.stopWaiters)
    assert.Equal(t, rightNow + int64(2 * time.Second), timer.ExpireAt())

    fc.Advance(2 * time.Second)

    assert.Equal(t, rightNow + int64(2 * time.Second), fc.RightNow())

    <- stop
    assert.Equal(t, 1, tn.foo)
    assert.Empty(t, timer.stopWaiters)

    wg.Wait()
}
