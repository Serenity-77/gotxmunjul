package logger

import (
    // "fmt"
    "testing"
    "github.com/stretchr/testify/assert"
)


func TestLogWrapper(t *testing.T) {
    logWrapper := NewLogWrapper(nil)
    logWrapper.Debugf("Hello %s", "World")
    logWrapper.Infof("Hello %s", "World")
    logWrapper.Printf("Hello %s", "World")
    logWrapper.Warnf("Hello %s", "World")
    logWrapper.Warningf("Hello %s", "World")
    logWrapper.Errorf("Hello %s", "World")
    logWrapper.Fatalf("Hello %s", "World")
    logWrapper.Panicf("Hello %s", "World")
}


type dummyOut struct {
    items   []string
}

func (do *dummyOut) Write(b []byte) (int, error) {
    do.items = append(do.items, string(b))
    return len(b), nil
}

type dummyFormatter struct {
    entries []*Entry
}

func (df *dummyFormatter) Format(entry *Entry) ([]byte, error) {
    df.entries = append(df.entries, entry)
    return []byte(entry.Message), nil
}

func TestLogWrapperNotNil(t *testing.T) {
    logger := New()
    logger.SetOutput(&dummyOut{})

    formatter := &dummyFormatter{}
    logger.SetFormatter(formatter)

    logger.SetLevel(DebugLevel)

    logWrapper := NewLogWrapper(logger)

    logWrapper.Debugf("Hello %s", "World")
    assert.Equal(t, DebugLevel, formatter.entries[0].Level)
    logWrapper.Infof("Hello %s", "World")
    assert.Equal(t, InfoLevel, formatter.entries[1].Level)
    logWrapper.Printf("Hello %s", "World")
    assert.Equal(t, InfoLevel, formatter.entries[2].Level)
    logWrapper.Warnf("Hello %s", "World")
    assert.Equal(t, WarnLevel, formatter.entries[3].Level)
    logWrapper.Warningf("Hello %s", "World")
    assert.Equal(t, WarnLevel, formatter.entries[4].Level)
    logWrapper.Errorf("Hello %s", "World")
    assert.Equal(t, ErrorLevel, formatter.entries[5].Level)
    logWrapper.Fatalf("Hello %s", "World")
    assert.Equal(t, FatalLevel, formatter.entries[6].Level)
    logWrapper.Panicf("Hello %s", "World")
    assert.Equal(t, PanicLevel, formatter.entries[7].Level)

}
