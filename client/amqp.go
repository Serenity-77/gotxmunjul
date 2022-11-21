package client

import (
    "time"
    "sync"
    "github.com/sirupsen/logrus"
    amqp "github.com/rabbitmq/amqp091-go"
    txLogger "github.com/serenity-77/bagudung/logger"
    txUtils "github.com/serenity-77/bagudung/utils"
)


var _ IAmqpConnection   = (*amqpConnection)(nil)
var _ IAmqpChannel      = (*amqp.Channel)(nil)

type IAmqpConnection interface {
    Channel()                       (IAmqpChannel, error)
    Close()                         error
    NotifyClose(chan *amqp.Error)   chan *amqp.Error
    GetClock()                      txUtils.IClock
}

type IAmqpChannel interface {
    QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
}

type AmqpClient struct {
    conn            IAmqpConnection
    mu              sync.Mutex
    dialUrl         string
    dialConfig      *amqp.Config
    dialFunc        AmqpDialFunc
    logger          txLogger.ILogger
    disconnect      chan struct{}
    loopStopped     chan struct{}
}

type AmqpDialFunc   func(string, *amqp.Config) (IAmqpConnection, error)

type amqpConnection struct {
    *amqp.Connection
    clock   txUtils.IClock
}

func (conn *amqpConnection) Channel() (IAmqpChannel, error) {
    return conn.Connection.Channel()
}

func (conn *amqpConnection) GetClock() txUtils.IClock {
    return conn.clock
}

func _defaultDialFunc(dialUrl string, dialConfig *amqp.Config) (IAmqpConnection, error) {
    var (
        conn *amqp.Connection
        err error
    )
    if dialConfig == nil {
        conn, err = amqp.Dial(dialUrl)
    } else {
        conn, err = amqp.DialConfig(dialUrl, *dialConfig)
    }

    return &amqpConnection{conn, txUtils.NewRealClock()}, err
}

func NewAmqpClientDialFunc(dialUrl string, dialConfig *amqp.Config, dialFunc AmqpDialFunc, logger txLogger.ILogger) (*AmqpClient, error) {
    client := &AmqpClient{
        dialUrl:        dialUrl,
        dialConfig:     dialConfig,
        dialFunc:       dialFunc,
        logger:         txLogger.NewLogWrapper(logger),
        loopStopped:    make(chan struct{}),
        disconnect:     make(chan struct{}),
    }

    if err := client.doConnect(); err != nil {
        client.logger = nil
        client.loopStopped = nil
        client.disconnect = nil
        if client.dialConfig != nil {
            client.dialConfig = nil
        }
        client = nil
        return client, err
    }

    go client.waitClose()

    return client, nil
}

func NewAmqpClient(dialUrl string, dialConfig *amqp.Config, logger ILogger) (*AmqpClient, error) {
    return NewAmqpClientDialFunc(dialUrl, dialConfig, _defaultDialFunc, logger)
}

func (c *AmqpClient) Channel() (IAmqpChannel, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.conn.Channel()
}

func (c *AmqpClient) Disconnect() error {
    c.mu.Lock()
    conn := c.conn
    c.mu.Unlock()

    err := conn.Close()

    close(c.disconnect)

    <- c.loopStopped

    c.conn = nil
    c.logger = nil
    c.loopStopped = nil
    c.disconnect = nil

    return err
}

func (c *AmqpClient) doConnect() error {
    if c.dialFunc == nil {
        c.dialFunc = _defaultDialFunc
    }
    if conn, err := c.dialFunc(c.dialUrl, c.dialConfig); err != nil {
        return err
    } else {
        c.mu.Lock()
        c.conn = conn
        c.mu.Unlock()
        return nil
    }
}

func (c *AmqpClient) waitClose() {
    defer close(c.loopStopped)

    for {
        closeChan := c.conn.NotifyClose(make(chan *amqp.Error))

        if reason, ok := <- closeChan; !ok {
            return
        } else {
            c.logger.Errorf("AmqpClient Disconnected: %#v", reason)

            connected := false
            reconnectInterval := 2

            reconnectTimer := c.conn.GetClock().Timer(time.Duration(reconnectInterval) * time.Second)

            for !connected {
                c.logger.Infof("Reconnecting AmqpClient in %d seconds", reconnectInterval)

                if reconnectInterval == 10 {
                    reconnectInterval = 0
                }

                select {
                case <- reconnectTimer.C:
                    if err := c.doConnect(); err != nil {
                        c.logger.Errorf("AmqpClient reconnecting error: %#v", err)
                        reconnectInterval += 2
                        reconnectTimer.Reset(time.Duration(reconnectInterval) * time.Second)
                    } else {
                        connected = true
                        reconnectTimer.Stop()
                        c.logger.Infof("AmqpClient Connected")
                    }
                case <- c.disconnect:
                    reconnectTimer.Stop()
                    return
                }
            }
        }
    }
}
