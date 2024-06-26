package oss

import (
	"errors"
	"io"
	"strings"
	"time"

	"github.com/DOVECYJ/phoenix/oss/ali"
)

var (
	ErrUnsupportSource = errors.New("un support source")
)

// oss客户端接口
type OssClient interface {
	OssWriter
	OssReader
	IsExist() (bool, error)
	SignUrl(d time.Duration) (string, error)
	Url() string
}

type OssWriter interface {
	Write(io.Reader) error
	WriteFile(string) error
}

type OssReader interface {
	Read() (io.ReadCloser, error)
	ReadToFile(dest string) error
}

// 打开oss客户端
func Open(source, filename string, overwrite bool) (OssClient, error) {
	source = strings.ToLower(source)
	switch source {
	case "ali":
		return ali.Open(filename, overwrite)
	case "tx":
		fallthrough
	default:
		return nil, ErrUnsupportSource
	}
}

// Copy Reader to oss
func Copy(dst OssWriter, src io.Reader) error {
	return dst.Write(src)
}

// duplicate OssReader to local dst
func Dump(dst io.Writer, src OssReader) error {
	r, err := src.Read()
	if err != nil {
		return err
	}
	defer r.Close()

	_, err = io.Copy(dst, r)
	return err
}
