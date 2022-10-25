package logger

import (
    "testing"
    "github.com/stretchr/testify/assert"
)


func TestBufferedLogWriter(t *testing.T) {
    w := &BufferedLoggerWriter{}
    w.Write([]byte("Hello"))
    w.Write([]byte("World"))
    expectedItems := [][]byte{[]byte("Hello"), []byte("World")}
    assert.Equal(t, expectedItems, w.AllItems())
}


func TestBufferedLogger(t *testing.T) {
    w := &BufferedLoggerWriter{}
    f := &NoLoggerFormatter{}

    logger, _ := CreateLogger(w, f, "info")

    logger.Info("Hello")
    logger.Info("World")
    logger.Infof("Hello %s", "World")

    expectedItems := [][]byte{[]byte("Hello"), []byte("World"), []byte("Hello World")}
    assert.Equal(t, expectedItems, w.AllItems())

    w.Clear()
    assert.Nil(t, w.AllItems())

    logger.Info("Hello")
    logger.Info("World")
    logger.Infof("Hello %s", "World")

    assert.Equal(t, expectedItems, w.AllItems())
}


func TestNullLogger(t *testing.T) {
    w := &NullLoggerWriter{}
    f := &NullLoggerFormatter{}
    logger, err := CreateLogger(w, f, "info")

    if err != nil {
        panic(err)
    }

    logger.Info("Hello")
    logger.Info("World")
    logger.Infof("Hello %s", "World")
}
