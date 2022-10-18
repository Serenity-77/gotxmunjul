package client

import (
    amqp "github.com/rabbitmq/amqp091-go"
)

type AmqpClient struct {
    conn        *amqp.Connection
    dialUrl     string
    dialConfig  *amqp.Config
}

var _dialFunc = func(dialUrl string, dialConfig *amqp.Config) (*amqp.Connection, error) {
    var (
        conn *amqp.Connection
        err error
    )
    if dialConfig == nil {
        conn, err = amqp.Dial(dialUrl)
    } else {
        conn, err = amqp.DialConfig(dialUrl, *dialConfig)
    }
    return conn, err
}

var _channelFunc = func(conn *amqp.Connection) (*amqp.Channel, error) {
    return nil, nil
}

func SetDialFunc(dialFunc func(string, *amqp.Config) (*amqp.Connection, error)) {
    _dialFunc = dialFunc
}

func SetChannelFunc(channelFunc func(*amqp.Connection) (*amqp.Channel, error)) {
    _channelFunc = channelFunc
}

func NewAmqpClient(dialUrl string, dialConfig *amqp.Config) (*AmqpClient, error) {
    client := &AmqpClient{}
    client.dialUrl = dialUrl
    client.dialConfig = dialConfig
    if err := client.doConnect(); err != nil {
        return nil, err
    }
    return client, nil
}


func (client *AmqpClient) GetChannel() (*amqp.Channel, error) {
    return _channelFunc(client.conn)
}

func (c *AmqpClient) doConnect() error {
    if conn, err := _dialFunc(c.dialUrl, c.dialConfig); err != nil {
        return err
    } else {
        c.conn = conn
        // closeChan := make(chan *amqp.Error)
        // closeChan = conn.NotifyClose(closeChan)
        // go c.onClose(closeChan)
        return nil
    }
}
