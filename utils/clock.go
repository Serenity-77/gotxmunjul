package utils

import (
    "time"
    "sort"
    "sync/atomic"
)

var _ ITimerImpl    = (*Timer)(nil)
var _ ITimerImpl    = (*FakeTimer)(nil)
var _ IClock        = (*FakeClock)(nil)

var _ ITimerImpl    = (*time.Timer)(nil)
var _ IClock        = (*RealClock)(nil)


type ITimerImpl interface {
    Reset(time.Duration)    bool
    Stop()                  bool
}

type IClock interface {
    Timer(time.Duration)    *Timer
}

type FakeClock struct {
    timers      []*FakeTimer
    rightNow    int64
    timerChan   chan *FakeTimer
    waitBlock   chan struct{}
    advance     chan advance
}

func NewFakeClock() *FakeClock {
    fc := &FakeClock{
        rightNow:   time.Now().UnixNano(),
        timerChan:  make(chan *FakeTimer),
        waitBlock:  make(chan struct{}),
        advance:    make(chan advance),
    }
    go fc.timerLoop()
    return fc
}

func (fc *FakeClock) Timer(duration time.Duration) *Timer {
    c := make(chan time.Time)
    ft := &FakeTimer{
        C:          c,
        c:          c,
        expireAt:   atomic.LoadInt64(&fc.rightNow) + int64(duration),
        fc:         fc,
    }
    timer := &Timer{ft, ft.C}
    fc.timerChan <- ft
    return timer
}

func (fc *FakeClock) RightNow() int64 {
    return atomic.LoadInt64(&fc.rightNow)
}

func (fc *FakeClock) timerLoop() {
    for {
        select {
        case timer := <- fc.timerChan:
            fc.timers = append(fc.timers, timer)
            fc.waitBlock <- struct{}{}
        case advance := <- fc.advance:
            fc.doAdvance(advance)
        }
    }
}

func (fc *FakeClock) WaitUntilBlock(n int) {
    for n > 0 {
        <- fc.waitBlock
        n--
    }
}

type advance struct {
    b           chan struct{}
    duration    time.Duration
}

func (fc *FakeClock) Advance(duration time.Duration) {
    advance := advance{
        b:          make(chan struct{}),
        duration:   duration,
    }
    fc.advance <- advance
    <- advance.b
}

func (fc *FakeClock) doAdvance(advance advance) {
    atomic.AddInt64(&fc.rightNow, int64(advance.duration))
    fc.sortTimers()
    for len(fc.timers) > 0 && fc.timers[0].expireAt <= fc.rightNow {
        timer := fc.timers[0]
        fc.timers = fc.timers[1:]
        timer.c <- time.Unix(0, int64(fc.rightNow))
    }
    advance.b <- struct{}{}
}

func (fc *FakeClock) sortTimers() {
    sort.Slice(fc.timers, func(i, j int) bool {
        if fc.timers[i].expireAt <= fc.timers[j].expireAt {
            return true
        }
        return false
    })
}

type FakeTimer struct {
    C           <- chan time.Time
    c           chan time.Time
    expireAt    int64
    fc          *FakeClock
}

func (ft *FakeTimer) Reset(d time.Duration) bool {
    ft.expireAt = ft.fc.RightNow() + int64(d)
    ft.fc.timerChan <- ft
    return true
}

func (ft *FakeTimer) Stop() bool {
    return true
}

type Timer struct {
    timer   ITimerImpl
    C       <- chan time.Time
}

func (t *Timer) Reset(duration time.Duration) bool {
    return t.timer.Reset(duration)
}

func (t *Timer) Stop() bool {
    return t.timer.Stop()
}


type RealClock struct{}

func NewRealClock() *RealClock {
    return &RealClock{}
}

func (rc *RealClock) Timer(d time.Duration) *Timer {
    rt := time.NewTimer(d)
    timer := &Timer{rt, rt.C}
    return timer
}
