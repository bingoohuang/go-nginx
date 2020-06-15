package nginxconf

import (
	"reflect"
	"sort"
	"strings"

	"github.com/bingoohuang/gonginx/directive"

	"github.com/bingoohuang/gou/str"
	"github.com/sirupsen/logrus"
)

type NginxServer struct {
	Listen    int
	Locations directive.Locations
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
	server.Locations = make([]directive.Location, 0)

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
