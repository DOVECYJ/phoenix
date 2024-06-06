package phoenix

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// The reponse for api application
type ApiResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

// error interface with a error code
type CodeError interface {
	Code() int // error code
	error
}

// Business error
type businessError struct {
	code int
	msg  string
}

func NewBusinessError(code int, err error) businessError {
	return businessError{code: code, msg: err.Error()}
}

func (b businessError) Code() int {
	return b.code
}

func (b businessError) Error() string {
	return b.msg
}

// Just int. It is used to quickly build CodeError.
type ErrCode int

func (c ErrCode) WithErr(err error) CodeError {
	return businessError{int(c), err.Error()}
}

func (c ErrCode) WithMsg(msg string) CodeError {
	return businessError{int(c), msg}
}

// Usage:
//
//	Code(10000).Err(err).Status(200).Render(w)
//	Code(10000).Msg("bulabula").Status(400).Render(w)
//	Data([]int{1,2,3}).Status(201).Render(w)
//
// It's a more flexible ApiResponse
type response map[string]any

// Set code for response
func (r response) Code(code int) response {
	r["code"] = code
	return r
}

// Set msg for response or and code if err is type of CodeError
func (r response) Err(err error) response {
	if cerr, ok := err.(CodeError); ok {
		r["code"] = cerr.Code()
		r["msg"] = cerr.Error()
	} else {
		r["msg"] = err.Error()
	}
	return r
}

// Set msg for response
func (r response) Msg(msg string) response {
	r["msg"] = msg
	return r
}

// Set status for http response
func (r response) Status(code int) response {
	r["status"] = code
	return r
}

// Set data for response
func (r response) Data(data any) response {
	r["data"] = data
	return r
}

// Set other key value for response
func (r response) KV(key string, val any) response {
	r[key] = val
	return r
}

// Send response to w
func (r response) Render(w http.ResponseWriter) {
	if status, ok := r["status"]; ok {
		w.WriteHeader(status.(int))
		delete(r, "status")
	}
	if err := json.NewEncoder(w).Encode(r); err != nil {
		slog.Error("render response", "err", err)
	}
}

// Return a response with code
func Code(code int) response {
	return response(map[string]any{
		"code": code,
		"msg":  "",
		"data": nil,
	})
}

// Return a response with msg or code if err is type of CodeError
func Err(err error) response {
	r := response(map[string]any{})
	if cerr, ok := err.(CodeError); ok {
		r["code"] = cerr.Code()
		r["msg"] = cerr.Error()
	} else {
		r["code"] = 0
		r["msg"] = err.Error()
	}
	r["data"] = nil
	return r
}

// Return a response with msg
func Msg(msg string) response {
	return response(map[string]any{
		"code": 0,
		"msg":  msg,
		"data": nil,
	})
}

// Return a response with status
func Status(code int) response {
	return response(map[string]any{"status": code})
}

// Retun a response with data
func Data(data any) response {
	return response(map[string]any{
		"code": 0,
		"msg":  "",
		"data": data,
	})
}
