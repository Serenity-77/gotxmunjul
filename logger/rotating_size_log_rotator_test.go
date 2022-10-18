package logger


import (
    "os"
    "time"
    "reflect"
    "testing"
    "fmt"
    "github.com/stretchr/testify/assert"
)


type fileInfo struct {}

func (f *fileInfo) Name() string {
    return ""
}

func (f *fileInfo) Size() int64 {
    return 256
}

func (f *fileInfo) Mode() os.FileMode {
    return os.FileMode(0755)
}

func (f *fileInfo) ModTime() time.Time {
    return time.Now()
}

func (f *fileInfo) IsDir() bool {
    return false
}

func (f *fileInfo) Sys() interface{} {
    return 0
}


func TestRotatingSizeLogFileNew(t *testing.T) {
    rotatingSizeLogRotator := RotatingSizeLogRotatorNew(100, 5)

    assert.Equal(t, int64(0), rotatingSizeLogRotator.size)
    assert.Equal(t, int64(100), rotatingSizeLogRotator.maxRotateSize)
    assert.Equal(t, 5, rotatingSizeLogRotator.maxRotateFiles)
    assert.True(t, rotatingSizeLogRotator.initial)
}


type mFileInfo struct {
    fileInfo
}

func (m *mFileInfo) Size() int64 {
    return 0
}


func TestRotatingSizeLogFileShouldRotate(t *testing.T) {
    oldLoggerFileStatFunc := loggerFileStatFunc

    defer func() {
        loggerFileStatFunc = oldLoggerFileStatFunc
    }()

    loggerFileStatFunc = func(name string) (os.FileInfo, error) {
        return &mFileInfo{}, nil
    }

    logFile := &LogFile{}
    rotatingSizeLogRotator := RotatingSizeLogRotatorNew(50, 5)

    assert.True(t, rotatingSizeLogRotator.initial)

    shouldRotate := rotatingSizeLogRotator.ShouldRotate(logFile, []byte("0123456789"))
    assert.False(t, shouldRotate)
    assert.False(t, rotatingSizeLogRotator.initial)

    for i := 0; i < 4; i++ {
        shouldRotate = rotatingSizeLogRotator.ShouldRotate(logFile, []byte("0123456789"))
        assert.False(t, shouldRotate)
    }

    shouldRotate = rotatingSizeLogRotator.ShouldRotate(logFile, []byte("0123456789"))

    assert.True(t, shouldRotate)
    assert.Equal(t, int64(50), rotatingSizeLogRotator.size)
}


func TestRotatingSizeLogFileRotate(t *testing.T) {
    oldLoggerFileStatFunc := loggerFileStatFunc
    oldLoggerFileRenameFunc := loggerFileRenameFunc
    oldLoggerFilePathGlobFunc := loggerFilePathGlobFunc
    oldLoggerFileRemoveFunc := loggerFileRemoveFunc

    defer func() {
        loggerFileStatFunc = oldLoggerFileStatFunc
        loggerFileRenameFunc = oldLoggerFileRenameFunc
        loggerFilePathGlobFunc = oldLoggerFilePathGlobFunc
        loggerFileRemoveFunc = oldLoggerFileRemoveFunc
    }()

    loggerFileStatFunc = func(name string) (os.FileInfo, error) {
        return &mFileInfo{}, nil
    }

    rotatedPaths := make([]string, 0)
    globFiles := make(map[string]int)
    loggerFileRenameFunc = func(oldpath, newpath string) error {
        rotatedPaths = append(rotatedPaths, newpath)
        globFiles[newpath] = 1
        return nil
    }

    getGlobFiles := func() []string {
        r := []string{}
        for k := range globFiles {
            r = append(r, k)
        }
        return r
    }

    loggerFilePathGlobFunc = func(path string) ([]string, error) {
        return getGlobFiles(), nil
    }

    removedPaths := make([]string, 0)
    loggerFileRemoveFunc = func(name string) error {
        removedPaths = append(removedPaths, name)
        return nil
    }

    logFilePath := "/foo/bar/test1.log"
    dummyLogFile := &DummyLogFile{}
    rotatingSizeLogRotator := RotatingSizeLogRotatorNew(50, 10)
    openFileFn := dummyOpenFileWithWriteCloser(dummyLogFile)
    fileIO, _ := openFileFn(logFilePath, 0755)

    logFile := &LogFile{fileIO: fileIO, path: logFilePath, rotator: rotatingSizeLogRotator, openFileFn: openFileFn}

    expectedRotatedPaths := make([]string, 0)
    expectedRemovedPaths := make([]string, 0)

    for i := 0; i < 200; i++ {
        doTriggerRotatingSequence(t, logFile, rotatingSizeLogRotator)
        j := i + 1
        if j >= rotatingSizeLogRotator.maxRotateFiles {
            j = rotatingSizeLogRotator.maxRotateFiles
        }
        for ;j > 0; j-- {
            expectedRotatedPaths = append(expectedRotatedPaths, fmt.Sprintf("/foo/bar/test1.log.%d", j))
        }
        assertRotatingSizeLogFileAfterRotated(t, dummyLogFile, rotatingSizeLogRotator, expectedRotatedPaths, rotatedPaths, 0)
        if i + 1 > rotatingSizeLogRotator.maxRotateFiles {
            expectedRemovedPaths = append(expectedRemovedPaths, fmt.Sprintf("/foo/bar/test1.log.%d", rotatingSizeLogRotator.maxRotateFiles))
        }
        assert.True(t, reflect.DeepEqual(expectedRemovedPaths, removedPaths))
    }
}


func doTriggerRotatingSequence(t *testing.T, logFile *LogFile, rotatingSizeLogRotator *RotatingSizeLogRotator) {
    assert.Equal(t, int64(0), rotatingSizeLogRotator.size)

    for i := 0; i < 5; i++ {
        logFile.Write([]byte("HelloWorld"))
    }

    assert.Equal(t, int64(50), rotatingSizeLogRotator.size)

    logFile.Write([]byte("HelloWorld")) // trigger rotate

    // this should be 0 because Rotate method
    // reset size to 0
    assert.Equal(t, int64(0), rotatingSizeLogRotator.size)
}


func assertRotatingSizeLogFileAfterRotated(
    t *testing.T,
    dummyLogFile *DummyLogFile,
    rotatingSizeLogRotator *RotatingSizeLogRotator,
    expectedRotatedPaths, rotatedPaths []string,
    expectedSize int64) {
    assert.True(t, reflect.DeepEqual(expectedRotatedPaths, rotatedPaths))
    assert.Equalf(t, uint32(1), dummyLogFile.fileOpened, "%d != 0", dummyLogFile.fileOpened)
    assert.Equal(t, int64(expectedSize), rotatingSizeLogRotator.size)
}
