// Plug is the the new router system. It is not complete yet, so do not use it.
package plug

type plug interface {
	Handle(*Conn)
}

type PlugFunc func(*Conn)

func (p PlugFunc) Handle(c *Conn) { p(c) }
