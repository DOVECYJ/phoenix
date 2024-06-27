package ali

import (
	"io"
	"os"
	"path"
	"time"

	"github.com/DOVECYJ/phoenix"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func init() {
	phoenix.AfterLoadCondig("ali", func() error {
		return opt.LoadKeyAndValide("oss.ali")
	})
}

// ali oss config
type AliOssOption struct {
	phoenix.Configer[AliOssOption] `validate:"-"`

	Endpoint     string `validate:"required"`
	AccessKey    string `validate:"required"`
	AccessSecret string `validate:"required"`
	Bucket       string `validate:"required"`
	Domain       string `validate:"required"`
}

var (
	ossClient *oss.Client // ali oss client
	ossBucket *oss.Bucket // ali oss bucket
	opt       AliOssOption
)

// get or create ali oss client
func getClient() (*oss.Client, error) {
	if ossClient == nil {
		var err error
		ossClient, err = oss.New(opt.Endpoint, opt.AccessKey, opt.AccessSecret)
		if err != nil {
			return nil, err
		}
	}
	return ossClient, nil
}

// get or create ali oss bucket
func getBucket(name string) (*oss.Bucket, error) {
	if ossBucket == nil {
		c, err := getClient()
		if err != nil {
			return nil, err
		}
		// new bucket
		ossBucket, err = c.Bucket(name)
		if err != nil {
			return nil, err
		}
	}
	return ossBucket, nil
}

// open an oss flie
func Open(filename string, overwrite bool) (*bucket, error) {
	if overwrite {
		return newBucket(filename)
	} else {
		return newBucket(filename, oss.ForbidOverWrite(true))
	}
}

func newBucket(filename string, opts ...oss.Option) (*bucket, error) {
	b, err := getBucket(opt.Bucket)
	if err != nil {
		return nil, err
	}
	bucket := &bucket{
		bucket:   b,
		domain:   opt.Domain,
		filename: filename,
		opts:     append([]oss.Option{}, opts...),
	}
	return bucket, err
}

type bucket struct {
	bucket   *oss.Bucket
	domain   string
	filename string
	opts     []oss.Option
}

func (b *bucket) Write(r io.Reader) error {
	err := b.bucket.PutObject(b.filename, r, b.opts...)
	if err == nil {
		return nil
	}
	switch err := err.(type) {
	case oss.ServiceError:
		if err.StatusCode == 409 && err.Code == "FileAlreadyExists" {
			return os.ErrExist
		}
	}
	return err
}

func (b *bucket) WriteFile(name string) error {
	err := b.bucket.PutObjectFromFile(b.filename, name, b.opts...)
	if err == nil {
		return nil
	}
	switch err := err.(type) {
	case oss.ServiceError:
		if err.StatusCode == 409 && err.Code == "FileAlreadyExists" {
			return os.ErrExist
		}
	}
	return err
}

func (b *bucket) Read() (io.ReadCloser, error) {
	r, err := b.bucket.GetObject(b.filename)
	if err == nil {
		return r, nil
	}
	switch err := err.(type) {
	case oss.ServiceError:
		if err.StatusCode == 404 && err.Code == "NoSuchKey" {
			return r, os.ErrNotExist
		}
	}
	return r, err
}

func (b *bucket) ReadToFile(dest string) error {
	err := b.bucket.GetObjectToFile(b.filename, dest)
	if err == nil {
		return nil
	}
	switch err := err.(type) {
	case oss.ServiceError:
		if err.StatusCode == 404 && err.Code == "NoSuchKey" {
			return os.ErrNotExist
		}
	}
	return err
}

func (b *bucket) IsExist() (bool, error) {
	return b.bucket.IsObjectExist(b.filename)
}

func (b *bucket) SignUrl(d time.Duration) (string, error) {
	return b.bucket.SignURL(b.filename, oss.HTTPGet, int64(d.Seconds()))
}

func (b *bucket) Url() string {
	return path.Join(b.domain, b.filename)
}
