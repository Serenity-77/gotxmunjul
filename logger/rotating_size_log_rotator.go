package logger


import (
    "os"
    "fmt"
    "sort"
    "strings"
    "strconv"
)


var _ ILogFileRotator = (*RotatingSizeLogRotator)(nil)

type RotatingSizeLogRotator struct {
    size            int64
    maxRotateSize   int64
    maxRotateFiles  int
    initial         bool
}

func RotatingSizeLogRotatorNew(maxRotateSize int64, maxRotateFiles int) *RotatingSizeLogRotator {
    if maxRotateSize == 0 {
        panic("maxRotateSize must be > 0")
    }
    return &RotatingSizeLogRotator{
        size: 0,
        maxRotateSize: maxRotateSize,
        maxRotateFiles: maxRotateFiles,
        initial: true,
    }
}

func (r *RotatingSizeLogRotator) setSize(path string) error {
    fileInfo, err := loggerFileStatFunc(path)
    if err != nil && !os.IsNotExist(err) {
        return err
    } else if err != nil {
        r.size = 0
    } else {
        r.size = fileInfo.Size()
    }
    return nil
}


func (r *RotatingSizeLogRotator) ShouldRotate(logFile *LogFile, b []byte) bool {
    if r.initial {
        r.setSize(logFile.GetPath())
        r.size += int64(len(b))
        r.initial = false
        return false
    }

    if r.size >= r.maxRotateSize {
        return true
    }

    r.size += int64(len(b))

    return false
}


func (r *RotatingSizeLogRotator) Rotate(logFile *LogFile) error {
    path := logFile.GetPath()
    logIdentifiers := getLogIdentifiers(path)

    for _, id := range(logIdentifiers) {
        if id >= r.maxRotateFiles {
            loggerFileRemoveFunc(fmt.Sprintf("%s.%d", path, id))
        } else {
            loggerFileRenameFunc(fmt.Sprintf("%s.%d", path, id), fmt.Sprintf("%s.%d", path, id + 1))
        }
    }

    loggerFileRenameFunc(path, fmt.Sprintf("%s.%d", path, 1))

    r.size = 0

    return nil
}


func getLogIdentifiers(path string) []int {
    identifiers := make([]int, 0)
    fileList, err := loggerFilePathGlobFunc(fmt.Sprintf("%s.[0-9]*", path))

    if err != nil {
        return nil
    }

    if len(fileList) == 0 {
        return identifiers
    }

    for _, filename := range(fileList) {
        id, _ := strconv.Atoi(filename[strings.LastIndex(filename, ".") + 1:])
        // todo: handle error returned by strconv.Atoi
        identifiers = append(identifiers, id)
    }

    sort.Sort(sort.Reverse(sort.IntSlice(identifiers)))

    return identifiers
}
