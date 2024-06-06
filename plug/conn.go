package plug

import "net/http"

// The most important abstract
type Conn struct {
	w   http.ResponseWriter
	r   *http.Request
	err error
}

func (c *Conn) Terminate() {

}

func (c *Conn) Render() {}
