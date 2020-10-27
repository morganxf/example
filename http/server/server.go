package server

import (
	"context"
	"net"
)

type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}

type Server struct {
	Addr string
}

func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

}

// serve 接收在Listener上的链接
func (srv *Server) serve(l net.Listener) error {
	ctx := context.Background()
	for {
		rw, err := l.Accept()
		if err != nil {
		}
		c := srv.newConn(rw)
		go c.
	}
}

func (srv *Server) newConn(rwc net.Conn) *conn {
	c := &conn{
		server: srv,
		rwc:    rwc,
	}
	return c
}
