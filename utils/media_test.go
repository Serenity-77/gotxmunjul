package libs

import (
    "fmt"
    "os"
    "bytes"
    "testing"
    "image"
    "path/filepath"
    "github.com/stretchr/testify/assert"
)

const testDataDir = "../testdata/media"

func TestCreateImageThumbailFromReader(t *testing.T) {
    doTestCreateImageThumbailFromReader(t, fmt.Sprintf("%s/nyx.png", testDataDir), "png")
    doTestCreateImageThumbailFromReader(t, fmt.Sprintf("%s/scylla.jpg", testDataDir), "jpeg")
}

func TestCreateImageThumbnail(t *testing.T) {
    doTestCreateImageThumbail(t, fmt.Sprintf("%s/nyx.png", testDataDir), "png")
    doTestCreateImageThumbail(t, fmt.Sprintf("%s/scylla.jpg", testDataDir), "jpeg")
}


func doTestCreateImageThumbailFromReader(t *testing.T, path string, format string) {
    fileIO, err := os.Open(path)

    defer fileIO.Close()

    if err != nil {
        panic(err)
    }

    result := &bytes.Buffer{}

    err = CreateImageThumbnailFromReader(fileIO, 100, 200, result)
    assert.Nil(t, err)

    _, frmat, err := image.DecodeConfig(result)

    if err != nil {
        panic(err)
    }

    assert.Equal(t, format, frmat)
}


func doTestCreateImageThumbail(t *testing.T, path string, format string) {
    result := &bytes.Buffer{}

    err := CreateImageThumbnail(path, 100, 200, result)
    assert.Nil(t, err)

    _, frmat, err := image.DecodeConfig(result)

    if err != nil {
        panic(err)
    }

    assert.Equal(t, format, frmat)
}

func TestFFMPEGVideoAdapter(t *testing.T) {
    adapter, err := FFMPEGVideoAdapterNew(filepath.Join(testDataDir, "big_buck_bunny.mp4"))

    if err != nil {
        panic(err)
    }

    assert.Equal(t, filepath.Join(testDataDir, "big_buck_bunny.mp4"), adapter.Path)
    assert.NotEqual(t, FFPROBFormat{}, adapter.Format)
    assert.Equal(t, filepath.Join(testDataDir, "big_buck_bunny.mp4"), adapter.Format.Filename)
}

func TestFFMPEGVideoAdapterFromReader(t *testing.T) {
    reader, err := os.Open(filepath.Join(testDataDir, "big_buck_bunny.mp4"))

    defer reader.Close()

    if err != nil {
        panic(err)
    }

    adapter, err := FFMPEGVideoAdapterNewFromReader(reader)

    if err != nil {
        panic(err)
    }

    defer os.Remove(adapter.Path)

    assert.NotNil(t, adapter.reader)
    assert.NotEmpty(t, adapter.Path)
    assert.NotEqual(t, FFPROBFormat{}, adapter.Format)
    assert.NotEmpty(t, adapter.Format.Filename)
}


func TestFFMPEGVideoAdapterSaveFrame(t *testing.T) {
    adapter, err := FFMPEGVideoAdapterNew(filepath.Join(testDataDir, "big_buck_bunny.mp4"))

    if err != nil {
        panic(err)
    }

    b := bytes.Buffer{}

    if err = adapter.SaveFrame(14, "jpg", &b); err != nil {
        panic(err)
    }

    assert.Greater(t, b.Len(), 0)
}

type testImageConf struct {
    image.Config
    format  string
}

func TestGetImageConfig(t *testing.T) {
    imageConfig := make(map[string]testImageConf)

    files := []string{
        filepath.Join(testDataDir, "nyx.png"),
        filepath.Join(testDataDir, "scylla.jpg"),
    }

    for _, file := range files {
        if fileIO, err := os.Open(file); err != nil {
            panic(err)
        } else if conf, format, err := image.DecodeConfig(fileIO); err != nil {
            panic(err)
        } else {
            defer fileIO.Close()
            imageConfig[file] = testImageConf{conf, format}
        }
    }

    for _, file := range files {
        if fileIO, err := os.Open(file); err != nil {
            panic(err)
        } else if conf, err := GetImageConfig(fileIO); err != nil {
            panic(err)
        } else {
            defer fileIO.Close()
            assert.Equal(t, conf.Width, imageConfig[file].Width)
            assert.Equal(t, conf.Height, imageConfig[file].Height)
            assert.Equal(t, conf.Format, imageConfig[file].format)
        }
    }
}
