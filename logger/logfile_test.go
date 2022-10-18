package logger


import (
    "os"
    "testing"
    "path"
    "io"
    "github.com/stretchr/testify/assert"
)



func TestLogFileNew(t *testing.T) {
    filename := "logfile1.log"
    cwd, err := os.Getwd()

    if err != nil {
        panic(err)
    }

    path := path.Join(cwd, filename)

    defer func() {
        os.Remove(path)
    }()

    logFile, err := LogFileNew(filename, cwd, 0755, nil)

    var _ io.WriteCloser = logFile.fileIO

    assert.Equal(t, filename, logFile.GetName())
    assert.Equal(t, cwd, logFile.GetDir())
    assert.Equal(t, path, logFile.GetPath())
    assert.Equal(t, os.FileMode(0755), logFile.GetPermission())
    assert.Nil(t, logFile.rotator)

    assert.FileExists(t, path)
}
