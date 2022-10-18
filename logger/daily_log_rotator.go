package logger

import (
    "time"
    "os"
    "fmt"
)

var _ ILogFileRotator = (*DailyLogFileRotator)(nil)


type DailyLogFileRotator struct {
    lastYear int
    lastMonth time.Month
    lastDay int
}


func DailyLogFileRotatorNew() *DailyLogFileRotator {
    rotator := &DailyLogFileRotator{}
    rotator.setLastDate()
    return rotator
}


func (d *DailyLogFileRotator) setLastDate() {
    now := loggerTimeNowFunc()
    d.lastYear = now.Year()
    d.lastMonth = now.Month()
    d.lastDay = now.Day()
}


func (d *DailyLogFileRotator) ShouldRotate(logFile *LogFile, b []byte) bool {
    now := loggerTimeNowFunc()
    return (now.Day() > d.lastDay) || (now.Month() > d.lastMonth) || (now.Year() > d.lastYear)
}


func (d *DailyLogFileRotator) Rotate(logFile *LogFile) error {
    oldPath := logFile.GetPath()
    newPath := fmt.Sprintf("%s.%d_%d_%d", oldPath, d.lastYear, d.lastMonth, d.lastDay)

    _, err := os.Stat(newPath)

    if err != nil && !os.IsNotExist(err) {
        // stat failed
        return err
    }

    if err != nil && os.IsNotExist(err) {
        if err := loggerFileRenameFunc(oldPath, newPath); err != nil {
            return err
        }
    }
    
    d.setLastDate()

    return nil
}
