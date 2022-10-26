package logger

import (
    "io"
    "os"
    "fmt"
    "syscall"
    "sync"
    "github.com/sirupsen/logrus"
)


func CreateLogger(out io.Writer, formatter logrus.Formatter, level string) (*logrus.Logger, error) {
    if formatter == nil {
        return nil, fmt.Errorf("formatter parameter cannot be nil")
    }

    logger := logrus.New()
    logger.SetOutput(out)
    logger.SetFormatter(formatter)

    if logLevel, err := logrus.ParseLevel(level); err != nil {
        return nil, err
    } else {
        logger.SetLevel(logLevel)
    }

    return logger, nil
}

func CreateFor(out io.Writer, prefix string, level string) (*logrus.Logger, error) {
    return CreateLogger(out, &TextFormatterWithPrefix{LogPrefix: prefix}, level)
}

func DuplicateWithFormatter(logger *logrus.Logger, formatter logrus.Formatter) *logrus.Logger {
    dupLogger := logrus.New()
    dupLogger.SetOutput(logger.Out)
    dupLogger.SetFormatter(formatter)
    dupLogger.SetLevel(logger.GetLevel())
    return dupLogger
}


func DuplicateWithLogPrefix(logger *logrus.Logger, prefix string) *logrus.Logger {
    formatter := &TextFormatterWithPrefix{}
    formatter.LogPrefix = prefix
    return DuplicateWithFormatter(logger, formatter)
}

func RedirectStandardIO(logger *logrus.Logger) {
    logFile := logger.Out.(*LogFile)

    stdoutFd := int(os.Stdout.Fd())
    stderrFd := int(os.Stderr.Fd())
    logFileFd := int(logFile.fileIO.(*os.File).Fd())

    syscall.Dup2(logFileFd, stdoutFd)
    syscall.Dup2(logFileFd, stderrFd)
}


type BufferedLoggerWriter struct {
    m       sync.Mutex
    items   [][]byte
}

func (bw *BufferedLoggerWriter) Write(b []byte) (int, error) {
    bw.m.Lock()
    defer bw.m.Unlock()
    bw.items = append(bw.items, b)
    return len(b), nil
}

func (bw *BufferedLoggerWriter) AllItems() [][]byte {
    bw.m.Lock()
    defer bw.m.Unlock()
    return bw.items
}

func (bw *BufferedLoggerWriter) Clear() {
    bw.m.Lock()
    defer bw.m.Unlock()
    bw.items = nil
}

type NoLoggerFormatter struct{}

func (nf *NoLoggerFormatter) Format(entry *logrus.Entry) ([]byte, error) {
    return []byte(entry.Message), nil
}


type NullLoggerWriter struct {}

func (nw *NullLoggerWriter) Write(b []byte) (int, error) {
    return len(b), nil
}

type NullLoggerFormatter struct {}

func (nf *NullLoggerFormatter) Format(entry *logrus.Entry) ([]byte, error) {
    return nil, nil
}
