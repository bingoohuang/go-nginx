package directive

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

// nolint:gochecknoinits
func init() {
	RegisterFactory(&proxyPassNaming{})
}

type proxyPassNaming struct{}

func (i proxyPassNaming) Create() Processor {
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
	proxyPath := ""

	if r.URL.Path == "" { // only host, 这时候 location 匹配的完整路径将直接透传给 url
		proxyPath = rq.URL.Path
	} else {
		proxyPath = strings.TrimPrefix(rq.URL.Path, r.LocationPath)
	}

	targetPath := util.TryPrepend(filepath.Join(r.URL.Path, proxyPath), "/")
	p := gonet.ReverseProxy(rq.URL.Path, r.URL.Host, targetPath, 10*time.Second) // nolint:gomnd
	p.ServeHTTP(w, rq)

	return ProcessTerminate
}
