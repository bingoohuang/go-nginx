package util

import (
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

type ResponseWriterLog struct {
	*http.Request
	w http.ResponseWriter

	status int
	head   http.Header
	body   string
}

func (r *ResponseWriterLog) WriteHeader(statusCode int) {
	r.w.WriteHeader(statusCode)
	r.status = statusCode
}

func (r *ResponseWriterLog) Header() http.Header { return r.head }
func (r *ResponseWriterLog) ContentType() string { return r.head.Get("Content-Type") }

func (r *ResponseWriterLog) Write(bytes []byte) (int, error) {
	switch r.ContentType() {
	case "application/json", "text/plain":
		r.body += string(bytes)
	}

	return r.w.Write(bytes)
}

func (r *ResponseWriterLog) LogResponse() {
	log.Printf("response Status: %d ContentType: %s Body: %s, Header:%v",
		r.status, r.ContentType(), r.body, r.head)
}

func WrapLog(w http.ResponseWriter, r *http.Request) *ResponseWriterLog {
	// Content-Type: multipart/form-data; boundary=------------------------7063eacb22d97aa0
	dumpRequestBody := !strings.Contains(r.Header.Get("Content-Type"), "multipart")

	dump, _ := httputil.DumpRequest(r, dumpRequestBody)
	log.Printf("request RemoteAddr: %s DumpRequest: %s", r.RemoteAddr, dump)

	return &ResponseWriterLog{Request: r, w: w, status: http.StatusOK, head: w.Header()}
}
