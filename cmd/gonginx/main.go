package main

import (
	"flag"
	"fmt"

	"github.com/bingoohuang/gonginx/nginxconf"
	"github.com/bingoohuang/gou/file"
	"github.com/sirupsen/logrus"
)

// nolint gochecknoglobals
var (
	configFile string
)

func main() {
	flag.StringVar(&configFile, "c", "conf/nginx.conf", "config file")
	flag.Parse()

	if err := file.SingleFileExists(configFile); err != nil {
		logrus.Fatalf("failed to find config file%s: %v", configFile, err)
	}

	conf, err := nginxconf.Parse(file.ReadBytes(configFile))

	if err != nil {
		logrus.Fatalf("failed to pare config file%s: %v", configFile, err)
	}

	fmt.Printf("config %+v\n", conf)

	servers := conf.ParseServers()
	if len(servers) == 0 {
		servers = append(servers, nginxconf.NginxServer{
			Listen: 8000, // nolint gomnd
			Locations: []nginxconf.Location{{
				Path: "/",
			}},
		})
	}

	for _, server := range servers {
		go server.Start()
	}

	select {}
}
