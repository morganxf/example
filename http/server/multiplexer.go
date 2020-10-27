package server

import (
	"net/http"
	"sync"
)

var (
	DefaultServeMux = &defaultServeMux
	defaultServeMux ServeMux
)

type ServeMux struct {
	mu sync.RWMutex
	m  map[string]muxEntry
	es []muxEntry
	// TODO: for what ?
	hosts bool
}

type muxEntry struct {
	h       Handler
	pattern string
}

func NewServeMux() *ServeMux {
	return new(ServeMux)
}

func (mux *ServeMux) ServeHTTP(w ResponseWriter, r *Request) {
	// TODO: * means ?
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			// https://yanni4night.github.io/http/2014/04/28/http-connection-header.html
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

func (mux *ServeMux) Handler(r *Request) (h Header, pattern string) {

}
