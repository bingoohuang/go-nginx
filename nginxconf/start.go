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

func (s NginxServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := s.Locations.FindLocation(r)

	if l == nil {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = fmt.Fprint(w, directive.WelcomeHTML)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}

		return
	}

	wr := util.WrapLog(w, r)
	defer wr.LogResponse()

	l.ServeHTTP(wr, r)
}
