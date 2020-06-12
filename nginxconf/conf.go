package nginxconf

import (
	"reflect"
	"sort"
	"strings"

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

type NginxServer struct {
	Listen    int
	Locations Locations
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
