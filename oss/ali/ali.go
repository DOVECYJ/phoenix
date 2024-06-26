package ali

import (
	"io"
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
	return b.bucket.PutObject(b.filename, r, b.opts...)
}

func (b *bucket) WriteFile(name string) error {
	return b.bucket.PutObjectFromFile(b.filename, name, b.opts...)
}

func (b *bucket) Read() (io.ReadCloser, error) {
	return b.bucket.GetObject(b.filename)
}

func (b *bucket) ReadToFile(dest string) error {
	return b.bucket.GetObjectToFile(b.filename, dest)
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
