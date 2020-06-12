package nginxconf

import (
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/bingoohuang/gonginx/util"

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
	// or like location ~* .(jpg|png|bmp).
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
	Seq                int              // 定义的顺序
	Priority           ModifierPriority // 匹配级别，从0开始，数字越小，匹配优先级越高
	Modifier           Modifier
	Path               string
	Index              string
	Root               string
	Alias              string
	ProxyPass          *url.URL
	ProxyPassLastSlash bool // 是否最后有/
	Echo               string
	Pattern            *regexp.Regexp
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

type Locations []Location

func (ls Locations) Len() int { return len(ls) }

func (ls Locations) Less(i, j int) bool {
	if ls[i].Priority < ls[j].Priority {
		return true
	}

	switch ls[i].Priority {
	case ModifierForward, ModifierNone:
		if len(ls[i].Path) > len(ls[j].Path) {
			return true
		}
	}

	return ls[i].Seq < ls[j].Seq
}

func (ls Locations) Swap(i, j int) { ls[i], ls[j] = ls[j], ls[i] }

type NginxServer struct {
	Listen    int
	Locations Locations
}

func (ls Locations) FindLocation(r *http.Request) *Location {
	path := r.URL.Path

	for _, l := range ls {
		switch l.Priority {
		case ModifierExactly:
			if l.Path == path {
				return &l
			}
		case ModifierForward:
			if path == l.Path || strings.HasPrefix(path, util.TryAppend(l.Path, "/")) {
				return &l
			}
		case ModifierRegular:
			if l.Pattern.FindString(path) != "" {
				return &l
			}
		case ModifierNone:
			if strings.HasPrefix(path, l.Path) {
				return &l
			}
		}
	}

	return nil
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
			l := parseLocation(block)
			l.Seq = len(server.Locations)
			server.Locations = append(server.Locations, l)
		default:
			logrus.Warnf("unsupported %+v", block)
		}
	}

	sort.Sort(server.Locations)

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
			proxyPass := block.Words[1]
			proxyPath, err := url.Parse(proxyPass)

			if err != nil {
				logrus.Fatalf("failed to parse proxy_pass %v", block.Words[1])
			}

			l.ProxyPass = proxyPath
			l.ProxyPassLastSlash = strings.HasSuffix(proxyPass, "/")

		case "echo":
			l.Echo = block.Words[1]
		default:
			logrus.Warnf("unsupported %+v", block.Words)
		}
	}

	return l
}
