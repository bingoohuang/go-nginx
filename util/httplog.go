package util

import (
	"net/http"
	"net/http/httputil"

	"github.com/sirupsen/logrus"
)

type ResponseWriterLog struct {
	w http.ResponseWriter

	status int
	head   http.Header
	body   string
}

func (r *ResponseWriterLog) WriteHeader(statusCode int) { r.w.WriteHeader(statusCode) }
func (r *ResponseWriterLog) Header() http.Header        { return r.head }
func (r *ResponseWriterLog) ContentType() string        { return r.head.Get("Content-Type") }

func (r *ResponseWriterLog) Write(bytes []byte) (int, error) {
	switch r.ContentType() {
	case "application/json", "text/plain":
		r.body += string(bytes)
	}

	return r.w.Write(bytes)
}

func (r *ResponseWriterLog) LogResponse() {
	logrus.Infof("response Status: %d ContentType: %s Body: %s", r.status, r.ContentType(), r.body)
}

func WrapLog(w http.ResponseWriter, r *http.Request) *ResponseWriterLog {
	dump, _ := httputil.DumpRequest(r, true)
	logrus.Infof("request RemoteAddr: %s DumpRequest: %s", r.RemoteAddr, dump)

	return &ResponseWriterLog{w: w, status: http.StatusOK, head: w.Header()}
}
