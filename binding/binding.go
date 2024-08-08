package binding

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/DOVECYJ/phoenix"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-rel/changeset/params"
)

const (
	defaultMaxBytes = 1 << 25 // 32M
)

// Provide base data bind functions for struct type T.
//
// Usage:
//
//	type UserReq struct {
//		Binder[UserReq]
//		Name string
//	}
//	var req UserReq
//	err := req.Bind(r)
//
// The binding implementation is provide by github.com/gin-gonic/gin.
type Binder[T any] struct {
}

// Bind data fot T from http.Request
func (b *Binder[T]) Bind(r *http.Request) error {
	return binding.
		Default(r.Method, filterFlags(requestHeader(r, "Content-Type"))).
		Bind(r, (*T)(unsafe.Pointer(b)))
}

// Bind header data for T from http.Request
func (b *Binder[T]) BindHeader(r *http.Request) error {
	return binding.Header.Bind(r, (*T)(unsafe.Pointer(b)))
}

// Bind query data for T from http.Request
func (b *Binder[T]) BindQuery(r *http.Request) error {
	return binding.Query.Bind(r, (*T)(unsafe.Pointer(b)))
}

// Bind post form data for T from http.Request
func (b *Binder[T]) BindPost(r *http.Request) error {
	return binding.FormPost.Bind(r, (*T)(unsafe.Pointer(b)))
}

// Bind form data for T from http.Request
func (b *Binder[T]) BindForm(r *http.Request) error {
	return binding.Form.Bind(r, (*T)(unsafe.Pointer(b)))
}

// Bind json data for T from http.Request
func (b *Binder[T]) BindJSON(r *http.Request) error {
	return binding.JSON.Bind(r, (*T)(unsafe.Pointer(b)))
}

// Bind xml data for T from http.Request
func (b *Binder[T]) BindXML(r *http.Request) error {
	return binding.XML.Bind(r, (*T)(unsafe.Pointer(b)))
}

// Bind yaml data for T from http.Request
func (b *Binder[T]) BindYAML(r *http.Request) error {
	return binding.YAML.Bind(r, (*T)(unsafe.Pointer(b)))
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

func requestHeader(r *http.Request, key string) string {
	return r.Header.Get(key)
}

func Bind(r *http.Request, obj any) error {
	return binding.
		Default(r.Method, filterFlags(requestHeader(r, "Content-Type"))).
		Bind(r, obj)
}

func BindHeader(r *http.Request, obj any) error {
	return binding.Header.Bind(r, obj)
}

func BindQuery(r *http.Request, obj any) error {
	return binding.Query.Bind(r, obj)
}

func BindPost(r *http.Request, obj any) error {
	return binding.FormPost.Bind(r, obj)
}

func BindForm(r *http.Request, obj any) error {
	return binding.Form.Bind(r, obj)
}

func BindJSON(r *http.Request, obj any) error {
	return binding.JSON.Bind(r, obj)
}

func BindXML(r *http.Request, obj any) error {
	return binding.XML.Bind(r, obj)
}

func BindYAML(r *http.Request, obj any) error {
	return binding.YAML.Bind(r, obj)
}

// Save file from http.Request.
//
// The default form file name is 'file' and you can specify it by
// From("my-file") option.
// The default save path is system temporary directory, use Dir("/data/temp")
// can change it.
// The default save file name is the upload file name, this can be changed
// by Name("a.txt").
// If you want to replace existed file, you can use Replace option.
//
// Usage:
//
//	BindFile(r, Form("my-file"), Name("a.txt"), Dir("/data"), Replace)
func BindFile(r *http.Request, opts ...option) (int64, error) {
	meta := fileMeta{
		formFileName: "file",
		destPath:     os.TempDir(),
	}
	for i := range opts {
		opts[i](&meta)
	}

	r.ParseMultipartForm(defaultMaxBytes)
	src, header, err := r.FormFile(meta.formFileName)
	if err != nil {
		return 0, err
	}
	defer src.Close()

	if meta.destFileName == "" {
		meta.destFileName = header.Filename
	}

	destFile := filepath.Join(meta.destPath, meta.destFileName)
	if !meta.replaceWhenExist && exist(destFile) {
		return 0, os.ErrExist
	}

	// 创建不存在的目录
	if err = os.MkdirAll(meta.destPath, os.ModePerm); err != nil {
		return 0, err
	}
	// 保存文件
	f, err := os.Create(destFile)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return io.Copy(f, src)
}

type fileMeta struct {
	formFileName     string
	destPath         string
	destFileName     string
	replaceWhenExist bool
}

type option func(*fileMeta)

func Form(name string) option {
	return func(fm *fileMeta) {
		fm.formFileName = name
	}
}

func Dir(dir string) option {
	return func(fm *fileMeta) {
		fm.destPath = dir
	}
}

func Name(name string) option {
	return func(fm *fileMeta) {
		fm.destFileName = name
	}
}

func Replace(fm *fileMeta) {
	fm.replaceWhenExist = true
}

func exist(path string) bool {
	_, err := os.Stat(path)
	return errors.Is(err, os.ErrNotExist)
}

func Attr(r *http.Request) (params.Params, error) {
	if r.Method == http.MethodGet {
		return FormAttr(r)
	}
	contentType := filterFlags(requestHeader(r, "Content-Type"))
	// TODO: support more content type
	switch contentType {
	case binding.MIMEJSON:
		return JsonAttr(r)
	case binding.MIMEMultipartPOSTForm:
		return MultipartFormAttr(r)
	default:
		return FormAttr(r)
	}
}

func FormAttr(r *http.Request) (params.Params, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	return params.ParseForm(r.Form), nil
}

func PostFromArrt(r *http.Request) (params.Params, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	return params.ParseForm(r.PostForm), nil
}

func QueryArrt(r *http.Request) (params.Params, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	return params.ParseForm(r.URL.Query()), nil
}

func MultipartFormAttr(r *http.Request) (params.Params, error) {
	if err := r.ParseMultipartForm(defaultMaxBytes); err != nil {
		return nil, err
	}
	return params.ParseForm(r.MultipartForm.Value), nil
}

func JsonAttr(r *http.Request) (params.Params, error) {
	if r == nil || r.Body == nil {
		return nil, errors.New("invalid request")
	}
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return params.ParseJSON(string(bs)), nil
}

// Get ID form r.Context() by key, the ID is int type.
func ContextID(r *http.Request) int {
	return r.Context().Value(phoenix.ID).(int)
}

// Get int valur form r.Context() by key.
func ContextInt(r *http.Request, key string) int {
	return r.Context().Value(phoenix.CtxKey(key)).(int)
}

// Get int64 valur form r.Context() by key.
func ContextInt64(r *http.Request, key string) int64 {
	return r.Context().Value(phoenix.CtxKey(key)).(int64)
}

// Get float32 valur form r.Context() by key.
func ContextFloat32(r *http.Request, key string) float32 {
	return r.Context().Value(phoenix.CtxKey(key)).(float32)
}

// Get float64 valur form r.Context() by key.
func ContextFloat64(r *http.Request, key string) float64 {
	return r.Context().Value(phoenix.CtxKey(key)).(float64)
}

// Get string valur form r.Context() by key.
func ContextString(r *http.Request, key string) string {
	return r.Context().Value(phoenix.CtxKey(key)).(string)
}

// Get value from r.Context() by key.
func ContextVal[T any](r *http.Request, key string) T {
	return r.Context().Value(phoenix.CtxKey(key)).(T)
}
