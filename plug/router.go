package plug

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const phoenixConn _key = "phoenix.conn"

type (
	_key      string
	routeFunc func(chi.Router)
)

// router   是一个分发器，本身也是一个Pluger，考虑分发如何实现
type Router struct {
	mux *chi.Mux
}

func (r *Router) Handle(c *Conn) {
	// dispatch router
	ctx := context.WithValue(c.r.Context(), phoenixConn, c)
	r.mux.ServeHTTP(c.w, c.r.WithContext(ctx))
}

func (r *Router) Scope(prefix string, p ...routeFunc) {
	r.mux.Route(prefix, func(r chi.Router) {
		for i := range p {
			p[i](r)
		}
	})
}

func PipeThrough(p ...plug) routeFunc {
	mids := make([]func(http.Handler) http.Handler, len(p))
	for i := range p {
		mids[i] = wrapMiddleware(p[i])
	}
	return func(r chi.Router) {
		r.Use(mids...)
	}
}

func Scope(path string, p ...routeFunc) routeFunc {
	return func(r chi.Router) {
		r.Route(path, func(r chi.Router) {
			for i := range p {
				p[i](r)
			}
		})
	}
}

func Get(path string, p plug) routeFunc {
	return func(r chi.Router) {
		r.Get(path, wrap(p))
	}
}

func Post(path string, p plug) routeFunc {
	return nil
}

func Resource(path string, c Controller, p ...routeFunc) routeFunc {
	return nil
}

func wrap(p plug) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn := r.Context().Value(phoenixConn).(*Conn)
		p.Handle(conn)
	}
}

func wrapFunc(fn PlugFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn := r.Context().Value(phoenixConn).(*Conn)
		fn(conn)
	}
}

func wrapMiddleware(p plug) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn := r.Context().Value(phoenixConn).(*Conn)
			p.Handle(conn)
			next.ServeHTTP(w, r)
		})
	}
}

func test() {
	var r Router
	r.Scope("/user",
		PipeThrough(PlugFunc(func(c *Conn) {})),
		Get("/", PlugFunc(func(c *Conn) {})),
		Scope("/post",
			Resource("/", struct{ Controller }{},
				Resource("/post", struct{ Controller }{}),
			),
		),
	)
}
