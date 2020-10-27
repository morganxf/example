package server

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type unsupportedTEError struct {
	err string
}

func (uste *unsupportedTEError) Error() string {
	return uste.err
}

type transferReader struct {
	// Input
	Header        Header
	StatusCode    int
	RequestMethod string
	ProtoMajor    int
	ProtoMinor    int
	// output
	Body             io.ReadCloser
	ContentLength    int64
	TransferEncoding []string
	Close            bool
	Trailer          Header
}

func (t *transferReader) protoAtLeast(m, n int) bool {
	return t.ProtoMajor > m || (t.ProtoMajor == m && t.ProtoMinor >= n)
}

func (t *transferReader) fixTransferEncoding() error {
	// TODO: Transfer-Encoding ?
	raw, found := t.Header["Transfer-Encoding"]
	if !found {
		return nil
	}
	// TODO: why?
	delete(t.Header, "Transfer-Encoding")

	// Issue 12785; ignore Transfer-Encoding on HTTP/1.0 requests.
	if !t.protoAtLeast(1, 1) {
		return nil
	}

	// TODO: ?
	encodings := strings.Split(raw[0], ",")
	te := make([]string, 0, len(encodings))
	for _, encoding := range encodings {
		encoding = strings.ToLower(strings.TrimSpace(encoding))
		// TODO: identity ?
		if encoding == "identity" {
			break
		}
		// TODO: chunked ?
		if encoding != "chunked" {
			return &unsupportedTEError{fmt.Sprintf("unsupported transfer encoding: %q", encoding)}
		}
		// te slice 扩大1
		te = te[:len(te)+1]
		te[len(te)-1] = encoding
	}
	if len(te) > 1 {
		return &badStringError{"too many transfer encodings", strings.Join(te, ",")}
	} else if len(te) == 1 {
		// TODO: ?
		delete(t.Header, "Content-Length")
		t.TransferEncoding = te
	}
	return nil
}

func readTransfer(msg interface{}, r *bufio.Reader) (err error) {
	// 初始化transferReader。method默认值是GET，主要是为适应下面的Response
	t := &transferReader{RequestMethod: "GET"}

	isResponse := false
	switch rr := msg.(type) {
	case *response:
	case *Request:
		t.Header = rr.Header
		t.RequestMethod = rr.Method
		t.ProtoMajor = rr.ProtoMajor
		t.ProtoMinor = rr.ProtoMinor
		// TODO: ?
		t.StatusCode = 200
		t.Close = rr.Close
	default:
		panic("unexpected error")
	}

	// 默认使用协议HTTP/1.1
	if t.ProtoMajor == 0 && t.ProtoMinor == 0 {
		t.ProtoMajor, t.ProtoMinor = 1, 1
	}

	// 根据HTTP协议处理Transfer-Encoding和Content-Length
	if err = t.fixTransferEncoding(); err != nil {
		return err
	}

}

func fixLength()
