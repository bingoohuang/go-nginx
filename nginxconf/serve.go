package nginxconf

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/bingoohuang/gou/file"
)

func (l Location) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// proxy_pass url 反向代理的坑 https://xuexb.github.io/learn-nginx/example/proxy_pass.html
	if l.ProxyPass != nil {
		l.ProxyPass.Do(w, r)

		return
	}

	if l.DefaultType != nil {
		l.DefaultType.Do(w, r)
	}

	if l.Return != nil {
		l.Return.Do(w, r)

		return
	}

	if l.Echo != nil {
		l.Echo.Do(w, r)

		return
	}

	if l.Index != "" {
		// http://nginx.org/en/docs/http/ngx_http_index_module.html
		// processes requests ending with the slash character (‘/’).
		if strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, filepath.Join(r.URL.Path, l.Index), http.StatusFound)

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
