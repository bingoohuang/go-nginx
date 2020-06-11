package nginxconf

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/bingoohuang/gonet"
	"github.com/bingoohuang/gou/str"
	"github.com/sirupsen/logrus"
)

// ModifierPriority defines the priority of the modifier.
// https://end0tknr.wordpress.com/2015/12/22/location-match-priority-in-nginx/
// https://artfulrobot.uk/blog/untangling-nginx-location-block-matching-algorithm
// https://blog.csdn.net/qq_15766181/article/details/72829672
// priority	| prefix       	                          | example
// 1        | = (exactly)	                          | location = /path
// 2        | ^~ (forward match)	                  | location = /image
// 3        | ~ or ~* (regular & case-(in)sensitive)  | location ~ /image/
// 4        | NONE (forward match)                    | location /image
type ModifierPriority int

// https://stackoverflow.com/questions/5238377/nginx-location-priority
const (
	// ModifierExactly like location = /path matches the query / only.
	ModifierExactly ModifierPriority = iota + 1
	// ModifierForward like location ^~ /path.
	// matches any query beginning with /path/ and halts searching,
	// so regular expressions will not be checked.
	ModifierForward
	// ModifierRegular like location ~ \.(gif|jpg|jpeg)$
	// ModifierRegular like location ~* .(jpg|png|bmp).
	ModifierRegular
	// ModifierNone means none modifier for the location.
	ModifierNone
)

// Modifier is the location modifier.
type Modifier string

// Priority returns the priority of the location matching.
func (m Modifier) Priority() ModifierPriority {
	switch m {
	case "=":
		return ModifierExactly
	case "^~":
		return ModifierForward
	case "~", "~*":
		return ModifierRegular
	default:
		return ModifierNone
	}
}

// Location is the location.
type Location struct {
	Priority  ModifierPriority // 匹配级别，从0开始，数字越小，匹配优先级越高
	Modifier  Modifier
	Path      string
	Index     string
	Root      string
	Alias     string
	ProxyPass *url.URL
	Echo      string
	Pattern   *regexp.Regexp
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
		return strings.HasPrefix(path, TryAppend(l.Path, "/"))
	case ModifierRegular:
		return l.Pattern.FindString(path) != ""
	default:
		return strings.HasPrefix(path, l.Path)
	}
}

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

	file := r.URL.Path
	if l.Root != "" {
		file = filepath.Join(l.Root, file)
	} else {
		file = strings.TrimPrefix(file, "/")
	}

	if strings.HasSuffix(file, "/index.html") {
		r.URL.Path = "avoid index.html redirect... in ServeHTTP"
	}

	http.ServeFile(w, r, file)
}

// TryAppend tries to append the given suffix if it does not exists in the s.
func TryAppend(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s
	}

	return s + "/"
}

type Locations []Location

type NginxServer struct {
	Listen    int
	Locations Locations
}

func (ls Locations) FindLocation(r *http.Request) *Location {
	path := r.URL.Path

	if l := ls.findByModifierExactly(path); l != nil {
		return l
	}

	if l := ls.findByModifierForward(path); l != nil {
		return l
	}

	if l := ls.findByModifierRegular(path); l != nil {
		return l
	}

	if l := ls.findByModifierNone(path); l != nil {
		return l
	}

	return nil
}

func (ls Locations) findByModifierNone(path string) (longest *Location) {
	for _, l := range ls {
		if l.Priority == ModifierNone && strings.HasPrefix(path, l.Path) {
			ll := l

			if longest == nil || len(longest.Path) < len(ll.Path) {
				longest = &ll
			}
		}
	}

	return longest
}

func (ls Locations) findByModifierRegular(path string) *Location {
	for _, l := range ls {
		if l.Priority == ModifierRegular && l.Pattern.FindString(path) != "" {
			return &l
		}
	}

	return nil
}

func (ls Locations) findByModifierExactly(path string) *Location {
	for _, l := range ls {
		if l.Priority == ModifierExactly && l.Path == path {
			return &l
		}
	}

	return nil
}

func (ls Locations) findByModifierForward(path string) (longest *Location) {
	for _, l := range ls {
		if l.Priority == ModifierForward &&
			(path == l.Path || strings.HasPrefix(path, TryAppend(l.Path, "/"))) {
			ll := l

			if longest == nil || len(longest.Path) < len(ll.Path) {
				longest = &ll
			}
		}
	}

	return longest
}

func (conf NginxConfigureBlock) ParseServers() []NginxServer {
	servers := make([]NginxServer, 0)

	for i := 0; i < len(conf); i++ {
		switch {
		case reflect.DeepEqual(conf[i].Words, []string{"server"}):
			servers = append(servers, parseServer(conf[i].Block))
		case reflect.DeepEqual(conf[i].Words, []string{"http"}):
			return conf[i].Block.ParseServers()
		default:
			logrus.Warnf("unsupported %+v", conf[i])
		}
	}

	return servers
}

func parseServer(conf NginxConfigureBlock) (server NginxServer) {
	server.Listen = 8000
	server.Locations = make([]Location, 0)

	for _, block := range conf {
		if len(block.Words) == 0 {
			continue
		}

		switch strings.ToLower(block.Words[0]) {
		case "listen":
			server.Listen = str.ParseInt(block.Words[1])
		case "location":
			server.Locations = append(server.Locations, parseLocation(block))
		default:
			logrus.Warnf("unsupported %+v", block)
		}
	}

	return server
}

func parseLocation(conf NginxConfigureCommand) (l Location) {
	if len(conf.Words) == 2 { // nolint gomnd
		l.Path = conf.Words[1]
	} else {
		l.Modifier = Modifier(conf.Words[1])
		l.Path = conf.Words[2]
	}

	if l.Priority = l.Modifier.Priority(); l.Priority == ModifierRegular {
		reg := l.Path
		if l.Modifier == "~*" {
			reg = "(?i)" + reg
		}

		l.Pattern = regexp.MustCompile(reg)
	}

	for _, block := range conf.Block {
		switch strings.ToLower(block.Words[0]) {
		case "index":
			l.Index = block.Words[1]
		case "root":
			l.Root = block.Words[1]
		case "alias":
			l.Alias = block.Words[1]
		case "proxy_pass":
			proxyPath, err := url.Parse(block.Words[1])
			if err != nil {
				logrus.Fatalf("failed to parse proxy_pass %v", block.Words[1])
			}

			l.ProxyPass = proxyPath

		case "echo":
			l.Echo = block.Words[1]
		default:
			logrus.Warnf("unsupported %+v", block.Words)
		}
	}

	return l
}
