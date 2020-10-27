package server

import (
	"bufio"
	"context"
	"net"
)

type conn struct {
	server *Server // server是该conn关联的server。不可修改，不为空
	rwc    net.Conn
	bufr *bufio.Reader
	bufw *bufio.Writer
}

// 为该连接提供服务
func (c *conn) serve(ctx context.Context) {

	// HTTP/1.x from here on
	for {
		w, err := c.
	}
}

func (c *conn) readRequest(ctx context.Context) (w *response, err error) {
	req, err := readRequest(c.bufr, keepHostHeader)
	if err != nil {
		return nil, err
	}

}
