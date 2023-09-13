package nginxconf

import (
	"net/http"

	"github.com/bingoohuang/gonginx/directive"
)

func (s NginxServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if l := s.Locations.FindLocation(r); l != nil {
		l.ServeHTTP(w, r)
		return
	}

	if r.URL.Path == "/" {
		directive.Welcome(w)
	}
}
