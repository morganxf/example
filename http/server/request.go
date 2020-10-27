package server

import (
	"bufio"
	"io"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

const keepHostHeader = false

type Request struct {
	Method        string
	URL           *url.URL
	Proto         string
	ProtoMajor    int
	ProtoMinor    int
	Header        Header
	Body          io.ReadCloser
	ContentLength int64
	Host          string
	Form          url.Values
	PostForm      url.Values
	RemoteAddr    string
	RequestURI    string
	Close         bool
}

func (r *Request) ProtoAtLeast(major, minor int) bool {
	return r.ProtoMajor > major || r.ProtoMajor == major && r.ProtoMinor > minor
}

// TODO: ?
func (r *Request) isH2Upgrade() bool {
	return r.Method == "PRI" && len(r.Header) == 0 && r.URL.Path == "*" && r.Proto == "HTTP/2.0"
}

func readRequest(b *bufio.Reader, deleteHostHeader bool) (req *Request, err error) {
	tp := newTextprotoReader(b)
	req = new(Request)

	// 读取第一行数据。根据HTTP协议，第一行格式为：GET /index/html HTTP/1.0
	var s string
	if s, err = tp.ReadLine(); err != nil {
		return nil, err
	}
	defer func() {
		// 释放资源
		putTextprotoReader(tp)
		// 重新处理返回的err
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()
	var ok bool
	req.Method, req.RequestURI, req.Proto, ok = parseRequestLine(s)
	if !ok {
	}
	if req.ProtoMajor, req.ProtoMinor, ok = ParseHTTPVersion(req.Proto); !ok {
	}

	rawurl := req.RequestURI
	if req.URL, err = url.ParseRequestURI(rawurl); err != nil {
		return nil, err
	}

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	req.Header = Header(mimeHeader)
	fixPragmaCacheControl(req.Header)

	req.Host = req.URL.Host
	if req.Host == "" {
		req.Host = req.Header.get("Host")
	}
	if deleteHostHeader {
		delete(req.Header, "Host")
	}

	req.Close = shouldClose(req.ProtoMajor, req.ProtoMinor, req.Header, false)
	// TODO: readTransfer
	// TODO: ?
	if req.isH2Upgrade() {
		// Because it's neither chunked, nor declared:
		req.ContentLength = -1
		// 阻止服务器继续处理这个连接，本处理完毕之后应该被关闭
		req.Close = true
	}

	return req, nil
}

// parseRequestLine解析请求中的第一行数据。根据HTTP协议，第一行格式为：GET /index/html HTTP/1.0
func parseRequestLine(line string) (method, requestURI, proto string, ok bool) {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	// request格式错误
	if s1 < 0 || s2 < 0 {
		return
	}
	s2 += s1 + 1
	return line[:s1], line[s1+1 : s2], line[s2+1:], true
}

func ParseHTTPVersion(vers string) (major, minor int, ok bool) {
	const Big = 1000000
	switch vers {
	case "HTTP/1.1":
		return 1, 1, true
	case "HTTP/1.0":
		return 1, 0, true
	}
	if !strings.HasPrefix(vers, "HTTP/") {
		return
	}
	dot := strings.Index(vers, ".")
	if dot < 0 {
		return
	}
	major, err := strconv.Atoi(vers[5:dot])
	if err != nil || major < 0 || major > Big {
		return 0, 0, false
	}
	minor, err = strconv.Atoi(vers[dot+1:])
	if err != nil || minor < 0 || minor > Big {
		return 0, 0, false
	}
	return major, minor, false
}

var textprotoReaderPool sync.Pool

func newTextprotoReader(br *bufio.Reader) *textproto.Reader {
	if v := textprotoReaderPool.Get(); v != nil {
		tr := v.(*textproto.Reader)
		tr.R = br
		return tr
	}
	return textproto.NewReader(br)
}

func putTextprotoReader(r *textproto.Reader) {
	// 释放textproto.Reader中的bufio.Reader
	r.R = nil
	textprotoReaderPool.Put(r)
}
