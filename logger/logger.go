package logger

import (
    "github.com/sirupsen/logrus"
)

func New() *Logger {
    return logrus.New()
}

type Logger     = logrus.Logger
type Entry      = logrus.Entry
type Level      = logrus.Level
type Fields     = logrus.Fields
type IFormatter = logrus.Formatter
type ILogger    = logrus.FieldLogger



const (
	PanicLevel Level       = logrus.PanicLevel
	FatalLevel             = logrus.FatalLevel
	ErrorLevel             = logrus.ErrorLevel
	WarnLevel              = logrus.WarnLevel
	InfoLevel              = logrus.InfoLevel
	DebugLevel             = logrus.DebugLevel
	TraceLevel             = logrus.TraceLevel
)

var _ ILogger = (*logWrapper)(nil)

type logWrapper struct {
    logger  ILogger
}


func (lw *logWrapper) WithField(key string, value interface{}) *Entry {
    if lw.logger != nil {
        return lw.logger.WithField(key, value)
    }
    return &Entry{}
}

func (lw *logWrapper) WithFields(fields Fields) *Entry {
    if lw.logger != nil {
        return lw.logger.WithFields(fields)
    }
    return &Entry{}
}

func (lw *logWrapper) WithError(err error) *Entry {
    if lw.logger != nil {
        return lw.logger.WithError(err)
    }
    return &Entry{}
}


func (lw *logWrapper) Debugf(format string, args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Debugf(format, args...)
    }
}

func (lw *logWrapper) Infof(format string, args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Infof(format, args...)
    }
}

func (lw *logWrapper) Printf(format string, args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Printf(format, args...)
    }
}

func (lw *logWrapper) Warnf(format string, args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Warnf(format, args...)
    }
}

func (lw *logWrapper) Warningf(format string, args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Warningf(format, args...)
    }
}

func (lw *logWrapper) Errorf(format string, args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Errorf(format, args...)
    }
}

func (lw *logWrapper) Fatalf(format string, args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Fatalf(format, args...)
    }
}

func (lw *logWrapper) Panicf(format string, args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Panicf(format, args...)
    }
}

func (lw *logWrapper) Debug(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Debug(args...)
    }
}

func (lw *logWrapper) Info(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Info(args...)
    }
}

func (lw *logWrapper) Print(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Println(args...)
    }
}

func (lw *logWrapper) Warn(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Warn(args...)
    }
}

func (lw *logWrapper) Warning(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Warning(args...)
    }
}

func (lw *logWrapper) Error(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Error(args...)
    }
}

func (lw *logWrapper) Fatal(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Fatal(args...)
    }
}

func (lw *logWrapper) Panic(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Panic(args...)
    }
}

func (lw *logWrapper) Debugln(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Debugln(args...)
    }
}

func (lw *logWrapper) Infoln(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Infoln(args...)
    }
}

func (lw *logWrapper) Println(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Println(args...)
    }
}

func (lw *logWrapper) Warnln(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Warnln(args...)
    }
}

func (lw *logWrapper) Warningln(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Warningln(args...)
    }
}

func (lw *logWrapper) Errorln(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Errorln(args...)
    }
}


func (lw *logWrapper) Fatalln(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Fatalln(args...)
    }
}

func (lw *logWrapper) Panicln(args ...interface{}) {
    if lw.logger != nil {
        lw.logger.Panicln(args...)
    }
}


func NewLogWrapper(logger ILogger) ILogger {
    return &logWrapper{logger}
}
