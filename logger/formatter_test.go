package logger


import (
    "fmt"
    "time"
    "testing"
    "github.com/stretchr/testify/assert"

    "github.com/sirupsen/logrus"
)


type FormatterWithPrefixWriter struct {
    message []byte
}


func (wf *FormatterWithPrefixWriter) Write(b []byte) (int, error) {
    wf.message = b
    return len(b), nil
}


func TestTextFormatterWithPrefixWithPrefix(t *testing.T) {
    doTestTextFormatterWithPrefix(t, "FooBar")
    doTestTextFormatterWithPrefix(t, "")
}


func doTestTextFormatterWithPrefix(t *testing.T, logPrefix string) {
    logger := logrus.New()

    out := &FormatterWithPrefixWriter{}
    logger.SetOutput(out)

    formatter := &TextFormatterWithPrefix{}
    logger.SetFormatter(formatter)

    var expectedOutput = ""

    if logPrefix != "" {
        formatter.LogPrefix = logPrefix
        expectedOutput = fmt.Sprintf("[%s] [FooBar:INFO] Hello World\n", time.Now().Format("2006-01-02 15:04:05"))
    } else {
        expectedOutput = fmt.Sprintf("[%s] [INFO] Hello World\n", time.Now().Format("2006-01-02 15:04:05"))
    }

    logger.Info("Hello World")

    assert.Equal(t, expectedOutput, string(out.message))
}
