package server

import "golang.org/x/net/http/httpguts"

func shouldClose(major, minor int, header Header, removeCloseHeader bool) bool {
	if major < 1 {
		return true
	}
	conv := header["Connection"]
	hasClose := httpguts.HeaderValuesContainsToken(conv, "close")
	if major == 1 && minor == 0 {
		return hasClose || !httpguts.HeaderValuesContainsToken(conv, "kee-alive")
	}
	if hasClose && removeCloseHeader {
		header.Del("Connection")
	}
	return hasClose
}
