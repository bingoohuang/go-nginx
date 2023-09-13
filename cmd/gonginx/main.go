package main

import (
	"flag"
	"log"

	_ "github.com/bingoohuang/godaemon/autoload"
	_ "github.com/bingoohuang/golog/pkg/autoload"
	"github.com/bingoohuang/gonginx/directive"
	"github.com/bingoohuang/gonginx/nginxconf"
	"github.com/bingoohuang/gou/file"
)

var configFile string

func main() {
	flag.StringVar(&configFile, "c", "conf/nginx.conf", "config file")
	flag.Parse()

	if err := file.SingleFileExists(configFile); err != nil {
		log.Fatalf("failed to find config file%s: %v", configFile, err)
	}

	conf, err := nginxconf.Parse(file.ReadBytes(configFile))
	if err != nil {
		log.Fatalf("failed to pare config file%s: %v", configFile, err)
	}

	servers := conf.ParseServers()
	if len(servers) == 0 {
		servers = append(servers, nginxconf.NginxServer{
			ListenPort: 8000,
			Locations: []directive.Location{{
				Path: "/",
			}},
		})
	}

	runningServers := nginxconf.NewRunningServers()
	for _, server := range servers {
		runningServers.Register(server)
	}

	runningServers.Start()
}
