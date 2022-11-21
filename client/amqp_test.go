package client

import (
    "time"
    "testing"
    "sync/atomic"
    "github.com/stretchr/testify/assert"
    "github.com/sirupsen/logrus"
    amqp "github.com/rabbitmq/amqp091-go"
    txUtils  "github.com/serenity-77/bagudung/utils"
)

const (
    _DIAL_URL_TEST = "amqp://guest_test:guest_test@host_test:5672"
)

func _checkConnType(t *testing.T, conn interface{}) bool {
    if _, ok := conn.(IAmqpConnection); !ok {
        t.Errorf("%T does not implement IAmqpConnection", conn)
        return false
    } else if _, ok := conn.(*amqpConnection); !ok {
        t.Errorf("%T is not of type *amqp.Connection", conn)
        return false
    }
    return true
}

func _checkChannelType(t *testing.T, channel interface{}) bool {
    if _, ok := channel.(IAmqpChannel); !ok {
        t.Errorf("%T does not implement IAmqpChannel", channel)
        return false
    }
    return true
}

type _dialFuncT struct {
    dialFuncCalled  bool
}

func (f *_dialFuncT) doDial (dialUrl string, dialConfig *amqp.Config) (IAmqpConnection, error) {
    f.dialFuncCalled = true
    return &amqpConnection{&amqp.Connection{}, nil}, nil
}

func TestAmqpBasicConnection(t *testing.T) {
    dialFunc := &_dialFuncT{}
    assert.False(t, dialFunc.dialFuncCalled)

    client, err := NewAmqpClientDialFunc(_DIAL_URL_TEST, &amqp.Config{}, dialFunc.doDial, nil)

    assert.NotNil(t, client.conn)
    assert.True(t, _checkConnType(t, client.conn))
    assert.NoError(t, err)
    assert.Equal(t, _DIAL_URL_TEST, client.dialUrl)
    assert.NotNil(t, client.dialConfig)
    assert.True(t, dialFunc.dialFuncCalled)
    assert.NotNil(t, client.logger)
}

type noOpLoggerFormatter struct{}

func (nf *noOpLoggerFormatter) Format(entry *logrus.Entry) ([]byte, error) {
    return nil, nil
}

var _ IAmqpConnection   = (*FakeAmqpConnection)(nil)
var _ IAmqpChannel      = (*FakeAmqpChannel)(nil)

type FakeAmqpConnection struct {
    closeChans  []chan *amqp.Error
    closeWaiter chan struct{}
    clock       *txUtils.FakeClock
    lastError   *amqp.Error
}

func NewFakeAmqpConnection() *FakeAmqpConnection {
    conn := &FakeAmqpConnection{
        closeWaiter:    make(chan struct{}),
        clock:          txUtils.NewFakeClock(),
    }
    return conn
}

func (fc *FakeAmqpConnection) Channel() (IAmqpChannel, error) {
    return &FakeAmqpChannel{}, nil
}

func (fc *FakeAmqpConnection) Close() error {
    for len(fc.closeChans) > 0 {
        closeChan := fc.closeChans[0]
        fc.closeChans = fc.closeChans[1:]
        close(closeChan)
    }
    return fc.lastError
}


func (fc *FakeAmqpConnection) NotifyClose(closeChan chan *amqp.Error) chan *amqp.Error {
    fc.closeChans = append(fc.closeChans, closeChan)
    fc.closeWaiter <- struct{}{}
    return closeChan
}

func (fc *FakeAmqpConnection) GetClock() txUtils.IClock {
    return fc.clock
}

func (fc *FakeAmqpConnection) WaitNotifyClose() {
    <- fc.closeWaiter
}

func (fc *FakeAmqpConnection) TriggerClose(reason *amqp.Error) {
    fc.lastError = reason
    for len(fc.closeChans) > 0 {
        closeChan := fc.closeChans[0]
        fc.closeChans = fc.closeChans[1:]
        closeChan <- reason
    }
}

func FakeAmqpDialFunc(dialUrl string, dialConfig *amqp.Config) (IAmqpConnection, error) {
    return NewFakeAmqpConnection(), nil
}

type FakeAmqpChannel struct {

}

func (c *FakeAmqpChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
    return amqp.Queue{}, nil
}

func TestAmqpClientChannel(t *testing.T) {
    client, err := NewAmqpClientDialFunc(_DIAL_URL_TEST, &amqp.Config{}, FakeAmqpDialFunc, nil)

    assert.NotNil(t, client)
    assert.Nil(t, err)

    channel, err := client.Channel()

    assert.NotNil(t, channel)
    assert.Nil(t, err)
}

type connectError struct {}

func (ce *connectError) Error() string {
    return "Connect Error"
}

func TestAmqpClientConnectError(t *testing.T) {
    connectError := &connectError{}

    dialFunc := func(dialUrl string, dialConfig *amqp.Config) (IAmqpConnection, error) {
        return nil, connectError
    }

    client, err := NewAmqpClientDialFunc(_DIAL_URL_TEST, &amqp.Config{}, dialFunc, nil)

    assert.Nil(t, client)
    assert.NotNil(t, err)
    assert.ErrorIs(t, err, connectError)
}

func TestAmqpClientReconnect(t *testing.T) {
    var reconnectError int32

    dialFn := func(dialUrl string, dialConfig *amqp.Config) (IAmqpConnection, error) {
        if atomic.LoadInt32(&reconnectError) == 0 {
            return FakeAmqpDialFunc(dialUrl, dialConfig)
        } else {
            return nil, &connectError{}
        }
    }

    client, _ := NewAmqpClientDialFunc(_DIAL_URL_TEST, nil, dialFn, nil)

    assert.NotNil(t, client)

    conn := client.conn.(*FakeAmqpConnection)
    conn.WaitNotifyClose()

    assert.Equal(t, 1, len(conn.closeChans))

    conn.TriggerClose(&amqp.Error{
        Code:       302,
        Reason:     "CONNECTION_FORCED - broker forced connection closure with reason 'shutdown'",
        Server:     true,
        Recover:    false,
    })

    assert.Empty(t, conn.closeChans)

    conn.clock.WaitUntilBlock(1)

    atomic.StoreInt32(&reconnectError, 1)
    reconnectTimer := conn.clock.GetTimer(0)
    timerStopped := reconnectTimer.WaitStop()

    intervals := []time.Duration{
        2 * time.Second,
        4 * time.Second,
        6 * time.Second,
        8 * time.Second,
        10 * time.Second,
        2 * time.Second,
        4 * time.Second,
    }

    for _, interval := range intervals {
        rightNow := conn.clock.RightNow()
        assert.Equal(t, rightNow + int64(interval), reconnectTimer.ExpireAt())
        conn.clock.Advance(interval)
        conn.clock.WaitUntilBlock(1)
        assert.Equal(t, rightNow + int64(interval), conn.clock.RightNow())
    }

    rightNow := conn.clock.RightNow()
    assert.Equal(t, rightNow + int64(6 * time.Second), reconnectTimer.ExpireAt())

    atomic.StoreInt32(&reconnectError, 0)

    conn.clock.Advance(6 * time.Second)
    <- timerStopped

    client.Channel()
    assert.Equal(t, rightNow + int64(6 * time.Second), conn.clock.RightNow())

    // check race
    for i := 0; i < 100; i++ {
        client.Channel()
    }

    assert.NotSame(t, client.conn, conn)
    client.conn.(*FakeAmqpConnection).WaitNotifyClose()
    assert.NotEmpty(t, client.conn.(*FakeAmqpConnection).closeChans)

    assert.True(t, reconnectTimer.Stopped())
    assert.NotNil(t, client.conn)

    // check race
    for i := 0; i < 100; i++ {
        client.Channel()
    }
}

func assertDisconnected(t *testing.T, client *AmqpClient, loopStopped chan struct{}, disconnect chan struct{}) {
    _, ok := <- loopStopped
    assert.False(t, ok)
    _, ok = <- disconnect
    assert.False(t, ok)

    assert.Nil(t, client.conn)
    assert.Nil(t, client.logger)
    assert.Nil(t, client.loopStopped)
    assert.Nil(t, client.disconnect)
}


func TestAmqpClientDisconnect(t *testing.T) {
    client, _ := NewAmqpClientDialFunc(_DIAL_URL_TEST, nil, FakeAmqpDialFunc, nil)

    assert.NotNil(t, client)

    loopStopped := client.loopStopped
    disconnect := client.disconnect

    client.conn.(*FakeAmqpConnection).WaitNotifyClose()

    closeChan := client.conn.(*FakeAmqpConnection).closeChans[0]

    err := client.Disconnect()

    assert.Nil(t, err)

    _, ok := <- closeChan
    assert.False(t, ok)
    assertDisconnected(t, client, loopStopped, disconnect)
}

func TestAmqpClientDisconnectWhileReconnecting(t *testing.T) {
    client, _ := NewAmqpClientDialFunc(_DIAL_URL_TEST, nil, FakeAmqpDialFunc, nil)

    assert.NotNil(t, client)

    loopStopped := client.loopStopped
    disconnect := client.disconnect

    conn := client.conn.(*FakeAmqpConnection)
    conn.WaitNotifyClose()

    assert.Equal(t, 1, len(conn.closeChans))

    conn.TriggerClose(&amqp.Error{
        Code:       302,
        Reason:     "CONNECTION_FORCED - broker forced connection closure with reason 'shutdown'",
        Server:     true,
        Recover:    false,
    })

    assert.Empty(t, conn.closeChans)

    conn.clock.WaitUntilBlock(1)

    reconnectTimer := conn.clock.GetTimer(0)

    rightNow := conn.clock.RightNow()

    assert.Equal(t, rightNow + int64(2 * time.Second), reconnectTimer.ExpireAt())

    err := client.Disconnect()

    assert.NotNil(t, err)

    assert.True(t, reconnectTimer.Stopped())

    assertDisconnected(t, client, loopStopped, disconnect)
}
