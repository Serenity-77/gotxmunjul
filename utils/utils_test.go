package libs

import (
    "os"
    "errors"
    "path/filepath"
    "testing"
    "github.com/stretchr/testify/assert"
)



func TestFileExistsFileFalse(t *testing.T) {
    isExists, err := FileExists("/some/non/existence/files.txt")
    assert.False(t, isExists)
    assert.Nil(t, err)
}

func TestFileExistsDirFalse(t *testing.T) {
    isExists, err := FileExists("/some/non/existence/dir")
    assert.False(t, isExists)
    assert.Nil(t, err)
}

func TestFileExistsFileTrue(t *testing.T) {
    path := filepath.Join("../testdata", "test_file_exists.txt")

    f, err := os.Create(path)

    if err != nil {
        panic(err)
    }

    defer func() {
        f.Close()
        if err := os.Remove(path); err != nil {
            panic(err)
        }
    }()

    isExists, err := FileExists(path)

    assert.True(t, isExists)
    assert.Nil(t, err)
}

func TestFileExistsDirTrue(t *testing.T) {
    path := filepath.Join("../testdata", "test_dir_exists")

    err := os.Mkdir(path, os.FileMode(0775))

    if err != nil {
        panic(err)
    }

    defer func() {
        if err := os.Remove(path); err != nil {
            panic(err)
        }
    }()

    isExists, err := FileExists(path)

    assert.True(t, isExists)
    assert.Nil(t, err)
}


func TestFileExistsError(t *testing.T) {
    path := filepath.Join("./testdata", "file.txt")
    oldStatFunc := osStatFunc

    defer func() {
        osStatFunc = oldStatFunc
    }()

    osStatFunc = func(path string) (os.FileInfo, error) {
        return nil, os.ErrPermission
    }

    isExists, err := FileExists(path)

    assert.False(t, isExists)
    assert.True(t, errors.Is(err, os.ErrPermission))
}
