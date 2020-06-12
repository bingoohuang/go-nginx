package nginxconf

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bingoohuang/gonet"
	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/gonginx/util"
	"github.com/bingoohuang/gou/str"
)

// Return means http://nginx.org/en/docs/http/ngx_http_rewrite_module.html#return
// Syntax:	return code [text];
type Return struct {
	Code int
	Text string
}

func (r Return) Do(w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(r.Code)

	if r.Text != "" {
		_, _ = fmt.Fprint(w, r.Text)
	}
}

func ReturnParse(param []string) *Return {
	r := &Return{
		Code: str.ParseInt(param[0]),
	}

	if len(param) > 1 {
		r.Text = param[1]
	}

	return r
}

type DefaultType struct {
	Value string
}

func (r DefaultType) Do(w http.ResponseWriter, rq *http.Request) {
	w.Header().Set("Content-Type", r.Value)
}

func DefaultTypeParse(param []string) *DefaultType {
	r := &DefaultType{
		Value: param[0],
	}

	return r
}

type ProxyPass struct {
	LocationPath string
	URL          *url.URL
}

func ProxyPassParse(locationPath string, param []string) *ProxyPass {
	r := &ProxyPass{LocationPath: locationPath}

	proxyPass := param[0]
	proxyPath, err := url.Parse(proxyPass)

	if err != nil {
		logrus.Fatalf("failed to parse proxy_pass %v", param)
	}

	r.URL = proxyPath

	return r
}

func (r ProxyPass) Do(w http.ResponseWriter, rq *http.Request) {
	proxyPath := strings.TrimPrefix(r.URL.Path, r.LocationPath)
	targetPath := util.TryPrepend(filepath.Join(r.URL.Path, proxyPath), "/")
	p := gonet.ReverseProxy(r.URL.Path, r.URL.Host, targetPath, 10*time.Second) // nolint gomnd
	p.ServeHTTP(w, rq)
}

type Echo struct {
	Value string
}

func EchoParse(_ string, param []string) *Echo {
	return &Echo{
		Value: param[0],
	}
}

func (r Echo) Do(w http.ResponseWriter, rq *http.Request) {
	echo := r.Value
	echo = strings.ReplaceAll(echo, "$request", rq.RequestURI)
	_, _ = fmt.Fprint(w, echo)
}

// Location is the location.
type Location struct {
	Seq         int              // 定义的顺序
	Priority    ModifierPriority // 匹配级别，从0开始，数字越小，匹配优先级越高
	Modifier    Modifier
	Path        string
	Index       string
	Root        string
	Alias       string
	ProxyPass   *ProxyPass
	Echo        *Echo
	Return      *Return
	Pattern     *regexp.Regexp
	DefaultType *DefaultType
}

func (l Location) Matches(p ModifierPriority, r *http.Request) bool {
	if p != l.Priority {
		return false
	}

	path := r.URL.Path

	switch p {
	case ModifierExactly:
		return l.Path == path
	case ModifierForward:
		return strings.HasPrefix(path, util.TryAppend(l.Path, "/"))
	case ModifierRegular:
		return l.Pattern.FindString(path) != ""
	default:
		return strings.HasPrefix(path, l.Path)
	}
}
