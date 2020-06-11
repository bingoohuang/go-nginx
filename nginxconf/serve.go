package nginxconf

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/bingoohuang/gonet"
	"github.com/bingoohuang/gou/file"
)

func (l Location) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if l.ProxyPass != nil {
		proxyPath := strings.TrimPrefix(r.URL.Path, l.Path)
		targetHost := l.ProxyPass.Host
		targetPath := filepath.Join(l.ProxyPass.Path, proxyPath)
		proxy := gonet.ReverseProxy(r.URL.Path, targetHost, targetPath, 10*time.Second) // nolint gomnd
		proxy.ServeHTTP(w, r)

		return
	}

	if l.Echo != "" {
		_, _ = fmt.Fprint(w, l.Echo)
		return
	}

	if l.Index != "" {
		// http://nginx.org/en/docs/http/ngx_http_index_module.html
		// processes requests ending with the slash character (‘/’).
		if strings.HasSuffix(r.URL.Path, "/") {
			redirectURL := filepath.Join(r.URL.Path, l.Index)
			http.Redirect(w, r, redirectURL, http.StatusFound)

			return
		}
	}

	serveFile := r.URL.Path

	switch {
	case l.Root != "":
		// http://nginx.org/en/docs/http/ngx_http_core_module.html#root
		serveFile = filepath.Join(l.Root, serveFile)
	case l.Alias != "":
		// http://nginx.org/en/docs/http/ngx_http_core_module.html#alias
		// location /i/ { alias /data/w3/images/; }
		// on request of “/i/top.gif”, the file /data/w3/images/top.gif will be sent.
		serveFile = filepath.Join(l.Alias, strings.TrimPrefix(r.URL.Path, l.Path))
	default:
		serveFile = strings.TrimPrefix(serveFile, "/")
	}

	if r.URL.Path == "/" && file.SingleFileExists(serveFile) != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, welcome)

		return
	}

	if strings.HasSuffix(serveFile, "/index.html") {
		r.URL.Path = "avoid index.html redirect... in ServeHTTP"
	}

	http.ServeFile(w, r, serveFile)
}
