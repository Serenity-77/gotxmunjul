package logger

import (
    "time"
    "testing"
    "reflect"
    "github.com/stretchr/testify/assert"
)


func TestDailyLogFileNew(t *testing.T) {
    oldTimeNowFunc := loggerTimeNowFunc

    defer func() {
        loggerTimeNowFunc = oldTimeNowFunc
    }()

    now := time.Date(2022, 3, 7, 23, 9, 10, 0, time.UTC)

    loggerTimeNowFunc = func() time.Time {
        return now
    }

    dailyRotator := DailyLogFileRotatorNew()

    assert.Equal(t, dailyRotator.lastYear, now.Year())
    assert.Equal(t, dailyRotator.lastMonth, now.Month())
    assert.Equal(t, dailyRotator.lastDay, now.Day())
}


func TestDailyShouldRotate(t *testing.T) {
    oldTimeNowFunc := loggerTimeNowFunc

    defer func() {
        loggerTimeNowFunc = oldTimeNowFunc
    }()

    now := time.Date(2022, 3, 7, 20, 9, 10, 0, time.UTC)

    loggerTimeNowFunc = func() time.Time {
        return now
    }

    dailyRotator := DailyLogFileRotatorNew()

    assert.Equal(t, dailyRotator.lastYear, now.Year())
    assert.Equal(t, dailyRotator.lastMonth, now.Month())
    assert.Equal(t, dailyRotator.lastDay, now.Day())

    logFile := &LogFile{}

    shouldRotate := dailyRotator.ShouldRotate(logFile, []byte("hello"))
    assert.False(t, shouldRotate)

    for i := 0; i < 3; i++ {
        now = now.Add(time.Hour)
        shouldRotate = dailyRotator.ShouldRotate(logFile, []byte("hello"))
        assert.False(t, shouldRotate)
    }

    now = now.Add(time.Hour)
    shouldRotate = dailyRotator.ShouldRotate(logFile, []byte("hello"))
    assert.True(t, shouldRotate)
}

func TestDailyLogFileRotate(t *testing.T) {
    oldTimeNowFunc := loggerTimeNowFunc
    oldRenameFunc := loggerFileRenameFunc

    defer func() {
        loggerTimeNowFunc = oldTimeNowFunc
        loggerFileRenameFunc = oldRenameFunc
    }()

    now := time.Date(2022, 3, 7, 20, 9, 10, 0, time.UTC)

    loggerTimeNowFunc = func() time.Time {
        return now
    }

    rotatedPaths := make([]string, 0)
    loggerFileRenameFunc = func(oldpath, newpath string) error {
        rotatedPaths = append(rotatedPaths, newpath)
        return nil
    }

    logFilePath := "/foo/bar/test1.log"

    dummyLogFile := &DummyLogFile{}
    dailyRotator := DailyLogFileRotatorNew()
    openFileFn := dummyOpenFileWithWriteCloser(dummyLogFile)
    fileIO, _ := openFileFn(logFilePath, 0755)

    logFile := &LogFile{fileIO: fileIO, path: logFilePath, rotator: dailyRotator, openFileFn: openFileFn}

    assert.Equal(t, dailyRotator.lastYear, now.Year())
    assert.Equal(t, dailyRotator.lastMonth, now.Month())
    assert.Equal(t, dailyRotator.lastDay, now.Day())

    logFile.Write([]byte("Foo Bar"))
    assert.Truef(t, len(rotatedPaths) == 0, "len(rotatedPaths) = %d", len(rotatedPaths))
    logFile.Write([]byte("Foo Bar"))
    assert.Truef(t, len(rotatedPaths) == 0, "len(rotatedPaths) = %d", len(rotatedPaths))

    for i := 0; i < 3; i++ {
        now = now.Add(time.Hour)
        logFile.Write([]byte("Foo Bar"))
        assert.Truef(t, len(rotatedPaths) == 0, "len(rotatedPaths) = %d", len(rotatedPaths))
    }

    expectedRotatedPaths := []string{"/foo/bar/test1.log.2022_3_7"}

    now = now.Add(time.Hour)
    logFile.Write([]byte("Foo Bar"))
    assertDailyLogFileAfterRotated(t, dummyLogFile, dailyRotator, expectedRotatedPaths, rotatedPaths, now)

    logFile.Write([]byte("Foo Bar"))
    assertDailyLogFileAfterRotated(t, dummyLogFile, dailyRotator, expectedRotatedPaths, rotatedPaths, now)

    for i := 0; i < 18; i++ {
        now = now.Add(time.Hour)
        logFile.Write([]byte("Foo Bar"))
        assertDailyLogFileAfterRotated(t, dummyLogFile, dailyRotator, expectedRotatedPaths, rotatedPaths, now)
    }

    now = now.Add(6 * time.Hour)
    logFile.Write([]byte("Foo Bar"))

    expectedRotatedPaths = append(expectedRotatedPaths, "/foo/bar/test1.log.2022_3_8")
    assertDailyLogFileAfterRotated(t, dummyLogFile, dailyRotator, expectedRotatedPaths, rotatedPaths, now)
}


func assertDailyLogFileAfterRotated(
    t *testing.T,
    dummyLogFile *DummyLogFile,
    dailyRotator *DailyLogFileRotator,
    expectedRotatedPaths, rotatedPaths []string,
    now time.Time) {
    assert.True(t, reflect.DeepEqual(expectedRotatedPaths, rotatedPaths))
    assert.Equalf(t, uint32(1), dummyLogFile.fileOpened, "%d != 0", dummyLogFile.fileOpened)
    assert.Equal(t, now.Year(), dailyRotator.lastYear)
    assert.Equal(t, now.Month(), dailyRotator.lastMonth)
    assert.Equal(t, now.Day(), dailyRotator.lastDay)
}
