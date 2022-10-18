package logger

import (
    "io"
    "os"
    "syscall"
    "github.com/sirupsen/logrus"
)


// TODO:
// type ILogger interface {
//
// }



func CreateLogger(out io.Writer, formatter logrus.Formatter, level string) *logrus.Logger {
    logger := logrus.New()
    logger.SetOutput(out)
    logger.SetFormatter(formatter)

    logLevel, _ := logrus.ParseLevel(level)
    logger.SetLevel(logLevel)

    return logger
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
