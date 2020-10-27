package server

type ResponseWriter interface {
	Header() Header
	WriteHeader(statusCode int)
}

type response struct {
}
