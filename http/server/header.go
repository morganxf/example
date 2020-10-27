package server

import "net/textproto"

type Header map[string][]string

func (h Header) Set(key, value string) {
	textproto.MIMEHeader(h).Set(key, value)
}

func (h Header) get(key string) string {
	if v := h[key]; len(v) > 0 {
		return v[0]
	}
	return ""
}

func (h Header) Del(key string) {
	textproto.MIMEHeader(h).Del(key)
}

func fixPragmaCacheControl(header Header) {
	if hp, ok := header["Pragma"]; ok && len(hp) > 0 && hp[0] == "no-cache" {
		if _, found := header["Cache-Control"]; !found {
			header["Cache-Control"] = []string{"no-cache"}
		}
	}
}
