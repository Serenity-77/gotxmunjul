package logger


import (
    "fmt"
    "strings"
    "github.com/sirupsen/logrus"
)

type TextFormatterWithPrefix struct {
    LogPrefix string
}


func (formatter *TextFormatterWithPrefix) Format(entry *logrus.Entry) ([]byte, error) {
    timeString := fmt.Sprintf("[%s]", entry.Time.Format("2006-01-02 15:04:05"))

    var infoString = ""
    if formatter.LogPrefix != "" {
        infoString = formatter.LogPrefix
    }

    if infoString != "" {
        infoString = fmt.Sprintf("[%s:%s]", infoString, strings.ToUpper(entry.Level.String()))
    } else {
        infoString = fmt.Sprintf("[%s]", strings.ToUpper(entry.Level.String()))
    }

    logMessage := fmt.Sprintf("%s %s %s\n", timeString, infoString, entry.Message)

    return []byte(logMessage), nil
}


type BgdLogFormatter struct {
    LogFormat   string
    TimeFormat  string
    LogPrefix   string
}

const (
    defaultLogFormat    = "[{logTime}] [{logLevel}] {logMessage}"
    defaultTimeFormat   = "2006-01-02 15:04:05"
)

func (f *BgdLogFormatter) Format(entry *Entry) ([]byte, error) {
    logFormat := f.LogFormat
    timeFormat := f.TimeFormat

    if len(logFormat) == 0 {
        logFormat = defaultLogFormat
    }

    if len(timeFormat) == 0 {
        timeFormat = defaultTimeFormat
    }

    args := []string{
        "{logTime}",
        entry.Time.Format(timeFormat),
        "{logLevel}",
        strings.ToUpper(entry.Level.String()),
        "{logMessage}",
        entry.Message,
        "{logPrefix}",
        f.LogPrefix,
    }

    replacer := strings.NewReplacer(args...)
    logString := replacer.Replace(logFormat)

    return []byte(logString), nil
}
