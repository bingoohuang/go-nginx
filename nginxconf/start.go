package nginxconf

import (
	"fmt"
	"net/http"

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
		w.WriteHeader(http.StatusNotFound)
		return
	}

	l.ServeHTTP(w, r)
}
