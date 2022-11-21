package logger


import (
    "time"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/sirupsen/logrus"
)



func TestBgdLogFormatterDefault(t *testing.T) {
    formatter := &BgdLogFormatter{}
    formatted, err := formatter.Format(&Entry{
        Message:    "Hello World",
        Time:       time.Date(2022, time.November, 15, 2, 35, 45, 0, time.UTC),
        Level:      logrus.InfoLevel,
    })

    assert.Nil(t, err)
    assert.Equal(t, "[2022-11-15 02:35:45] [INFO] Hello World\n", string(formatted))
}

func TestBgdLogFormatterCustomFormat(t *testing.T) {
    formatter := &BgdLogFormatter{
        LogFormat:  "{logLevel} {logTime} {logMessage}",
    }

    formatted, err := formatter.Format(&Entry{
        Message:    "Hello World",
        Time:       time.Date(2022, time.November, 15, 2, 35, 45, 0, time.UTC),
        Level:      logrus.InfoLevel,
    })

    assert.Nil(t, err)
    assert.Equal(t, "INFO 2022-11-15 02:35:45 Hello World\n", string(formatted))
}

func TestBgdLogFormatterWithPrefix(t *testing.T) {
    formatter := &BgdLogFormatter{
        LogFormat:  "{logLevel} {logTime} {logPrefix} {logMessage}",
        LogPrefix:  "Hello-There",
    }

    formatted, err := formatter.Format(&Entry{
        Message:    "Hello World",
        Time:       time.Date(2022, time.November, 15, 2, 35, 45, 0, time.UTC),
        Level:      logrus.InfoLevel,
    })

    assert.Nil(t, err)
    assert.Equal(t, "INFO 2022-11-15 02:35:45 Hello-There Hello World\n", string(formatted))
}


func TestBgdLogFormatterEmptyPrefix(t *testing.T) {
    formatter := &BgdLogFormatter{
        LogFormat:  "{logLevel} {logTime} {logPrefix} {logMessage}",
    }

    formatted, err := formatter.Format(&Entry{
        Message:    "Hello World",
        Time:       time.Date(2022, time.November, 15, 2, 35, 45, 0, time.UTC),
        Level:      logrus.InfoLevel,
    })

    assert.Nil(t, err)
    assert.Equal(t, "INFO 2022-11-15 02:35:45  Hello World\n", string(formatted))
}
