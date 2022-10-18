package logger

/*
    https://github.com/twisted/twisted/blob/trunk/src/twisted/python/logfile.py
*/


import (
    "io"
    "os"
    "sync"
    "path/filepath"
)


type LogFile struct {
    fileIO          io.WriteCloser
    name, directory, path string
    defaultPerm     os.FileMode
    rotator         ILogFileRotator
    openFileFn      func(string, os.FileMode) (io.WriteCloser, error)
    writeLock       sync.Mutex
}


func LogFileNew(name, directory string, defaultPerm os.FileMode, rotator ILogFileRotator) (*LogFile, error) {
    if defaultPerm == 0 {
        defaultPerm = 0644
    }

    logFile := &LogFile{
        name: name,
        directory: directory,
        path: filepath.Join(directory, name),
        defaultPerm: defaultPerm,
    }

    logFile.openFileFn = openFile

    if err := logFile.openFile(); err != nil {
        return nil, err
    }

    logFile.rotator = rotator

    return logFile, nil
}

func (logFile *LogFile) openFile() error {
    fileIO, err := logFile.openFileFn(logFile.path, logFile.defaultPerm)

    if err != nil {
        return err
    }

    if logFile.fileIO != nil {
        if err = logFile.fileIO.Close(); err != nil {
            fileIO.Close()
            return err
        }
        logFile.fileIO = nil
    }

    logFile.fileIO = fileIO
    return nil
}

func openFile(path string, perm os.FileMode) (io.WriteCloser, error) {
    file, err := os.OpenFile(
        path,
        os.O_WRONLY | os.O_CREATE | os.O_APPEND,
        perm)

    if err != nil {
        return nil, err
    }

    return file, nil
}


func (logFile *LogFile) GetDir() string {
    return logFile.directory
}

func (logFile *LogFile) GetName() string {
    return logFile.name
}

func (logFile *LogFile) GetPath() string {
    return logFile.path
}

func (logFile *LogFile) GetPermission() os.FileMode {
    return logFile.defaultPerm
}

func (logFile *LogFile) Write(b []byte) (int, error) {
    if logFile.rotator != nil {
        logFile.writeLock.Lock()
        if logFile.rotator.ShouldRotate(logFile, b) {
            logFile.rotator.Rotate(logFile)     // for now ignore the errors
            logFile.openFile()
        }
        logFile.writeLock.Unlock()
    }

    return logFile.fileIO.Write(b)
}


type ILogFileRotator interface {
    ShouldRotate(*LogFile, []byte)   bool
    Rotate(*LogFile)         error
}
