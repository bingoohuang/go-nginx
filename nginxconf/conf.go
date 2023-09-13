package nginxconf

import (
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/bingoohuang/gonginx/directive"
	"github.com/bingoohuang/gou/str"
)

type NginxServer struct {
	ListenPort int
	Locations  directive.Locations
	ServerName string
}

func (conf NginxConfigureBlock) ParseServers() []NginxServer {
	servers := make([]NginxServer, 0)

	for i := 0; i < len(conf); i++ {
		words := conf[i].Words
		switch {
		case reflect.DeepEqual(words, []string{"server"}):
			servers = append(servers, parseServer(conf[i].Block))
		case reflect.DeepEqual(words, []string{"http"}):
			return conf[i].Block.ParseServers()
		default:
			log.Printf("W! unsupported %+v", conf[i])
		}
	}

	return servers
}

func parseServer(conf NginxConfigureBlock) (server NginxServer) {
	server.ListenPort = 8000
	server.Locations = make([]directive.Location, 0)

	for _, block := range conf {
		if len(block.Words) == 0 {
			continue
		}

		switch strings.ToLower(block.Words[0]) {
		case "listen":
			server.ListenPort = str.ParseInt(block.Words[1])
		case "server_name":
			server.ServerName = block.Words[1]
		case "location":
			l := parseLocation(block)
			l.Seq = len(server.Locations)
			server.Locations = append(server.Locations, l)
		default:
			log.Printf("W! unsupported %+v", block)
		}
	}

	sort.Sort(server.Locations)

	return server
}
