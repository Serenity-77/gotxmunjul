package utils

import (
    "fmt"
    "os"
    "io"
    "image"
    "strings"
    "bytes"
    "path/filepath"
    "os/exec"
    "image/jpeg"
    "image/png"
    "encoding/json"
    "github.com/nfnt/resize"
)


func CreateImageThumbnail(path string, width, height uint, result io.Writer) error {
    fileIO, err := os.Open(path)

    defer fileIO.Close()

    if err != nil {
        return err
    }

    return CreateImageThumbnailFromReader(fileIO, width, height, result)
}

func CreateImageThumbnailFromReader(reader io.Reader, width, height uint, result io.Writer) error {
    img, format, err := image.Decode(reader)

    if err != nil {
        return err
    }

    thumbnailImage := resize.Thumbnail(width, height, img, resize.Lanczos3)

    if format == "png" {
        err = png.Encode(result, thumbnailImage)
    } else {
        err = jpeg.Encode(result, thumbnailImage, nil)
    }

    if err != nil {
        return err
    }

    return nil
}

type ImageConfig struct {
    image.Config
    Format  string
}

func GetImageConfig(reader io.Reader) (*ImageConfig, error) {
    var (
        conf    image.Config
        format  string
        err     error
    )

    conf, format, err = image.DecodeConfig(reader)

    if seekable, ok := reader.(io.Seeker); ok {
        if _, err := seekable.Seek(0, 0); err != nil {
            return nil, err
        }
    }

    if err != nil {
        return nil, err
    }

    return &ImageConfig{conf, format}, nil
}

type FFPROBFormat struct {
    Filename string         `json:"filename"`
    NbStreams int           `json:"nb_streams"`
    FormatName string       `json:"format_name"`
    FormatLongName string   `json:"format_long_name"`
    StartTime string        `json:"start_time"`
    Duration string         `json:"duration"`
    BitRate string          `json:"bit_rate"`
    ProbeScore int          `json:"probe_score"`
    Tags struct {
        MajorBrand string       `json:"major_brand"`
        MinorVersion string     `json:"minor_version"`
        CompatibleBrands string `json:"compatible_brands"`
    }   `json:"tags"`
}


type FFMPEGVideoAdapter struct {
    Path string
    reader io.Reader
    Format FFPROBFormat
}


var (
    ffprobeArgs = strings.Split("-v error -print_format json -show_format", " ")
)

const (
    ffprobeCommand = "ffprobe"
    ffmpegCommand = "ffmpeg"
    ffmpegFrameArgs = "-y -v error -i %s -ss %d -vframes 1 %s"
)

func FFMPEGVideoAdapterNew(path string) (*FFMPEGVideoAdapter, error) {
    adapter := &FFMPEGVideoAdapter{Path: path}

    if _, err := exec.LookPath(ffprobeCommand); err != nil {
        return nil, fmt.Errorf("%s Command Not Found: %w", ffprobeCommand, err)
    }

    if _, err := exec.LookPath(ffmpegCommand); err != nil {
        return nil, fmt.Errorf("%s Command Not Found: %w", ffmpegCommand, err)
    }

    if err := adapter.runFFProbeInfo(); err != nil {
        return nil, err
    }

    return adapter, nil
}


func FFMPEGVideoAdapterNewFromReader(reader io.Reader) (*FFMPEGVideoAdapter, error) {
    path := filepath.Join("/tmp/", RandomString(7))

    fileIO, err := os.OpenFile(path, os.O_CREATE | os.O_WRONLY, os.FileMode(0755))

    defer fileIO.Close()

    if err != nil {
        return nil, err
    }

    if _, err = io.Copy(fileIO, reader); err != nil {
        return nil, err
    }

    adapter, err := FFMPEGVideoAdapterNew(path)

    if err != nil {
        return nil, err
    }

    adapter.reader = reader

    return adapter, nil
}


func (adapter *FFMPEGVideoAdapter) runFFProbeInfo() error {
    args := append(ffprobeArgs, adapter.Path)

    out := bytes.Buffer{}

    if err := adapter.runCmd(ffprobeCommand, args, &out); err != nil {
        return err
    }

    if err := json.Unmarshal(out.Bytes(), adapter); err != nil {
        return err
    }

    return nil
}

func (adapter *FFMPEGVideoAdapter) runCmd(command string, args []string, out io.Writer) error {
    cmd := exec.Command(command, args...)

    stdoutIO, err := cmd.StdoutPipe()
    if err != nil {
        return err
    }

    stderrIO, err := cmd.StderrPipe()
    if err != nil {
        return err
    }

    if err = cmd.Start(); err != nil {
        return err
    }

    if out != nil {
        stdoutBytes, err := io.ReadAll(stdoutIO)

        if err != nil {
            stdoutIO.Close()
            return fmt.Errorf("Failed to read ffprobe stdout: %#v", err)
        }

        if _, err = out.Write(stdoutBytes); err != nil {
            return err
        }
    }

    stderrBytes, err := io.ReadAll(stderrIO)
    if err != nil {
        stdoutIO.Close()
        return fmt.Errorf("Failed to read ffprobe stderr: %#v", err)
    }

    if err = cmd.Wait(); err != nil {
        cmdErr := err.(*exec.ExitError)
        exitCode := cmdErr.ProcessState.ExitCode()
        return fmt.Errorf(
                "%s command exiting with error code: %d, stderr: %s",
                command, exitCode, stderrBytes)
    }

    return nil
}

func (adapter *FFMPEGVideoAdapter) SaveFrame(duration int, format string, out io.Writer) error {
    if format != "jpeg" && format != "jpg" && format != "png" {
        return fmt.Errorf("Format %s not allowed", format)
    }

    if format == "jpeg" {
        format = "jpg"
    }

    outputPath := fmt.Sprintf("%s_Frame.%s", adapter.Path, format)
    args := strings.Split(fmt.Sprintf(ffmpegFrameArgs, adapter.Path, duration, outputPath), " ")

    if err := adapter.runCmd("ffmpeg", args, nil); err != nil {
        return err
    }

    defer os.Remove(outputPath)

    outputIO, err := os.Open(outputPath)

    if err != nil {
        return err
    }

    defer outputIO.Close()

    if err != nil {
        return err
    }

    if _, err = io.Copy(out, outputIO); err != nil {
        return err
    }

    return nil
}
