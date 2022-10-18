package logger

import (
    "time"
    "os"
    "io"
    "path/filepath"
)

var (
    loggerTimeNowFunc = time.Now
    loggerFileRenameFunc = os.Rename
    loggerFileStatFunc = os.Stat
    loggerFilePathGlobFunc = filepath.Glob
    loggerFileRemoveFunc = os.Remove
)


type DummyLogFile struct {
    fileOpened uint32
}


func (d *DummyLogFile) Write(b []byte) (int, error) {
    return len(b), nil
}

func (d *DummyLogFile) Close() error {
    d.fileOpened--
    return nil
}


func dummyOpenFile(path string, perm os.FileMode) (io.WriteCloser, error) {
    return &DummyLogFile{}, nil
}

func dummyOpenFileWithWriteCloser(r *DummyLogFile) func(string, os.FileMode) (io.WriteCloser, error) {
    return func(string, os.FileMode) (io.WriteCloser, error) {
        r.fileOpened++
        return r, nil
    }
}
