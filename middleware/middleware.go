package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/DOVECYJ/phoenix"
	"github.com/go-chi/chi/v5"
)

// MethodSpoofing allows to spoof PUT, PATCH and DELETE methods from HTML forms, using the _method field.
// <input type="hidden" name="_method" value="PUT">
func MethodSpoofing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			switch method := strings.ToUpper(r.PostFormValue("_method")); method {
			case http.MethodPut, http.MethodPatch, http.MethodDelete:
				r.Method = method
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Fetch id param in url
func FetchID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			slog.Warn("fetch id", "error", err)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), phoenix.ID, id)))
	})
}

// Fetch id with name in url
func FetchIDName(name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := strconv.Atoi(chi.URLParam(r, name))
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), phoenix.CtxKey(name), id)))
		})
	}
}
