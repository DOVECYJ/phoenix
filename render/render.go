package render

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/DOVECYJ/phoenix"
	"github.com/a-h/templ"
)

type option interface {
	apply(http.ResponseWriter)
}

// Status is an option that set http status code.
type Status int

func (s Status) apply(w http.ResponseWriter) {
	w.WriteHeader(int(s))
}

// Cookie is an option that write http cookie.
type Cookie http.Cookie

func (c *Cookie) apply(w http.ResponseWriter) {
	http.SetCookie(w, (*http.Cookie)(c))
}

type render interface {
	Render(w http.ResponseWriter) error
}

// Render data to w with opts.
func Render(w http.ResponseWriter, data any) {
	switch data := data.(type) {
	default:
		JSON(w, data)
	case phoenix.CodeError:
		JSON(w, phoenix.ApiResponse{Code: data.Code(), Msg: data.Error()})
	case templ.Component:
		HTML(w, data)
	case string:
		String(w, data)
	case []byte:
		Bytes(w, data)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		prinltln(w, data)
	case render:
		handleError(w, data.Render(w))
	case error:
		http.Error(w, data.Error(), http.StatusInternalServerError)
	}
}

// Render HTML component to w.
func HTML(w http.ResponseWriter, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(context.Background(), w); err != nil {
		handleError(w, err)
	}
}

// Render data to w in json format.
func JSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		handleError(w, err)
	}
}

func ApiData(w http.ResponseWriter, data any) {
	JSON(w, phoenix.ApiResponse{Data: data})
}

// Render code and error to w use ApiReponse.
func ApiError(w http.ResponseWriter, code int, err error) {
	JSON(w, phoenix.ApiResponse{Code: code, Msg: err.Error()})
}

// Send byte slice to w.
func Bytes(w http.ResponseWriter, data []byte) {
	// w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(data)
}

// Send string to w.
func String(w http.ResponseWriter, data string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, data)
}

func prinltln(w http.ResponseWriter, args ...any) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, args...)
}

func HttpStatus(w http.ResponseWriter, code int) http.ResponseWriter {
	w.WriteHeader(code)
	return w
}

func handleError(w http.ResponseWriter, err error) {
	if err != nil {
		slog.Error("render", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
