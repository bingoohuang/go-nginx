package nginxconf

import (
	"fmt"
	"net/http"

	"github.com/bingoohuang/gonginx/directive"

	"github.com/bingoohuang/gonginx/util"

	"github.com/sirupsen/logrus"
)

func (s NginxServer) Start() {
	mux := http.NewServeMux()
	mux.Handle("/", s)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", s.Listen),
		Handler: mux,
	}

	logrus.Infof("listening on %v", s.Listen)

	if err := server.ListenAndServe(); err != nil {
		logrus.Warnf("ListenAndServe error: %v", err)
	}
}

func (s NginxServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	w := util.WrapLog(rw, r)
	defer w.LogResponse()

	if l := s.Locations.FindLocation(r); l != nil {
		l.ServeHTTP(w, r)
		return
	}

	if r.URL.Path == "/" {
		directive.Welcome(w)
	}
}
