package plug

import "net/http"

var (
	defaultEndpoint = &Endpoint{}
)

func GetEndpoint() *Endpoint {
	return defaultEndpoint
}

func Plug(p ...plug) {
	defaultEndpoint.Plug(p...)
}

// endpoint 是一个http.Handler,有一个Pluger列表
type Endpoint struct {
	plugs []plug
}

func (e *Endpoint) Plug(p ...plug) {
	e.plugs = append(e.plugs, p...)
}

func (e *Endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn := &Conn{w: w, r: r}
	for i := range e.plugs {
		e.plugs[i].Handle(conn)
		if conn.err != nil {
			return
		}
	}
}
