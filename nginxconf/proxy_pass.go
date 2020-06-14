package nginxconf

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/bingoohuang/gonet"
	"github.com/bingoohuang/gonginx/util"
	"github.com/sirupsen/logrus"
)

// proxy_pass url 反向代理的坑 https://xuexb.github.io/learn-nginx/example/proxy_pass.html

// nolint gochecknoinits
func init() {
	AppendLocationFactory(&proxyPassNaming{})
}

type proxyPassNaming struct{}

func (i proxyPassNaming) Create() LocationProcessor {
	return &proxyPass{proxyPassNaming: i}
}

func (proxyPassNaming) Name() map[string]bool {
	return map[string]bool{
		"proxy_pass": true,
	}
}

type proxyPass struct {
	proxyPassNaming

	LocationPath string
	URL          *url.URL
}

func (r *proxyPass) GetProcessSeq() ProcessSeq {
	return Terminate
}

func (r *proxyPass) Parse(path string, name string, params []string) error {
	r.LocationPath = path

	proxyPass := params[0]
	proxyPath, err := url.Parse(proxyPass)

	if err != nil {
		logrus.Fatalf("failed to parse proxy_pass %v", proxyPass)
	}

	r.URL = proxyPath

	return nil
}

func (r *proxyPass) Do(l Location, w http.ResponseWriter, rq *http.Request) ProcessResult {
	proxyPath := strings.TrimPrefix(r.URL.Path, r.LocationPath)
	targetPath := util.TryPrepend(filepath.Join(r.URL.Path, proxyPath), "/")
	p := gonet.ReverseProxy(r.URL.Path, r.URL.Host, targetPath, 10*time.Second) // nolint gomnd
	p.ServeHTTP(w, rq)

	return ProcessTerminate
}
