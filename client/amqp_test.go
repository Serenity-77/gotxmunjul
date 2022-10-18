package client

import (
    "testing"
    "github.com/stretchr/testify/assert"
    amqp "github.com/rabbitmq/amqp091-go"
)

const (
    _DIAL_URL_TEST = "amqp://guest_test:guest_test@host_test:5672"
)

func _checkConnType(client interface{}) bool {
    _, ok := client.(*amqp.Connection)
    return ok
}

func _checkChannelType(channel interface{}) bool {
    _, ok := channel.(*amqp.Channel)
    return ok
}

type _dialFuncT struct {
    dialFuncCalled  bool
}

func (f *_dialFuncT) doDial (dialUrl string, dialConfig *amqp.Config) (*amqp.Connection, error) {
    f.dialFuncCalled = true
    return &amqp.Connection{}, nil
}

type _channelFuncT struct {
    channelFuncCalled   bool
}


func (f *_channelFuncT) getChannel(conn *amqp.Connection) (*amqp.Channel, error) {
    f.channelFuncCalled = true
    return &amqp.Channel{}, nil
}


func TestBasicWithSetDialFuncAndChannelFunc(t *testing.T) {
    dialFunc := &_dialFuncT{}
    assert.False(t, dialFunc.dialFuncCalled)

    SetDialFunc(dialFunc.doDial)

    client, err := NewAmqpClient(_DIAL_URL_TEST, &amqp.Config{})

    assert.NotNil(t, client.conn)
    assert.True(t, _checkConnType(client.conn))
    assert.NoError(t, err)
    assert.Equal(t, _DIAL_URL_TEST, client.dialUrl)
    assert.NotNil(t, client.dialConfig)
    assert.True(t, dialFunc.dialFuncCalled)

    channelFunc := &_channelFuncT{}
    assert.False(t, channelFunc.channelFuncCalled)

    SetChannelFunc(channelFunc.getChannel)

    channel, err := client.GetChannel()
    assert.NoError(t, err)
    assert.True(t, _checkChannelType(channel))
    assert.True(t, channelFunc.channelFuncCalled)
}
