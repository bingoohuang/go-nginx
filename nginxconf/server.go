package nginxconf

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"

	"go.elara.ws/pcre"
)

type container struct {
	dispatch map[string]http.Handler
	// serverNames 保留 server_name 的加入顺序
	serverNames  []string
	starStarting []string
	starEnding   []string
}

func (c *container) Register(server NginxServer) {
	if server.ServerName == "" {
		server.ServerName = "default_server"
	}
	c.dispatch[server.ServerName] = server
	c.serverNames = append(c.serverNames, server.ServerName)
}

// ServeHTTP server HTTP server
func (c *container) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Server names are defined using the server_name directive and determine which server block is used for a given request. See also “How nginx processes a request”. They may be defined using exact names, wildcard names, or regular expressions:
	//
	// server {
	//     listen       80;
	//     server_name  example.org  www.example.org;
	//     ...
	// }
	//
	// server {
	//     listen       80;
	//     server_name  *.example.org;
	//     ...
	// }
	//
	// server {
	//     listen       80;
	//     server_name  mail.*;
	//     ...
	// }
	//
	// server {
	//     listen       80;
	//     server_name  ~^(?<user>.+)\.example\.net$;
	//     ...
	// }
	// When searching for a virtual server by name, if name matches more than one of the specified variants,
	// e.g. both wildcard name and regular expression match, the first matching variant will be chosen,
	// in the following order of precedence:
	//
	// exact name
	// longest wildcard name starting with an asterisk, e.g. “*.example.org”
	// longest wildcard name ending with an asterisk, e.g. “mail.*”
	// first matching regular expression (in order of appearance in a configuration file)
	host := r.Host
	if h, _, err := net.SplitHostPort(r.Host); err == nil {
		host = h
	}

	if c.exactMatch(host, w, r) {
		return
	}

	if c.wildcardStartingMatch(host, w, r) {
		return
	}

	if c.regexMatchInOrderMatch(host, w, r) {
		return
	}

	c.defaultHandle(w, r)
}

func (c *container) prepare() {
	for serverName := range c.dispatch {
		if strings.HasPrefix(serverName, "*") {
			c.starStarting = append(c.starStarting, serverName)
		} else if strings.HasSuffix(serverName, "*") {
			c.starEnding = append(c.starEnding, serverName)
		}
	}

	sort.Slice(c.starStarting, func(i, j int) bool {
		return len(c.starStarting[i]) > len(c.starStarting[j])
	})
	sort.Slice(c.starEnding, func(i, j int) bool {
		return len(c.starEnding[i]) > len(c.starEnding[j])
	})
}

func (c *container) exactMatch(host string, w http.ResponseWriter, r *http.Request) bool {
	if h, ok := c.dispatch[host]; ok {
		h.ServeHTTP(w, r)
		return true
	}

	return false
}

func (c *container) wildcardStartingMatch(host string, w http.ResponseWriter, r *http.Request) bool {
	for _, start := range c.starStarting {
		if strings.HasSuffix(host, start[1:]) {
			c.dispatch[start].ServeHTTP(w, r)
			return true
		}
	}
	for _, end := range c.starEnding {
		if strings.HasPrefix(host, end[:len(end)-1]) {
			c.dispatch[end].ServeHTTP(w, r)
			return true
		}
	}

	return false
}

func (c *container) regexMatchInOrderMatch(host string, w http.ResponseWriter, r *http.Request) bool {
	for _, name := range c.serverNames {
		p, err := pcre.Compile(name)
		if err == nil {
			matched := p.MatchString(host)
			_ = p.Close()
			if matched {
				c.dispatch[name].ServeHTTP(w, r)
				return true
			}
		}
	}

	return false
}

func (c *container) defaultHandle(w http.ResponseWriter, r *http.Request) {
	if server, ok := c.dispatch["default_server"]; ok {
		server.ServeHTTP(w, r)
		return
	}

	c.dispatch[c.serverNames[0]].ServeHTTP(w, r)
}

type RunningServers struct {
	Servers map[int]*container
}

func NewRunningServers() *RunningServers {
	return &RunningServers{
		Servers: make(map[int]*container),
	}
}

func (s *RunningServers) Register(server NginxServer) {
	c, ok := s.Servers[server.ListenPort]
	if !ok {
		c = &container{dispatch: make(map[string]http.Handler)}
		s.Servers[server.ListenPort] = c
	}
	c.Register(server)
}

func (s *RunningServers) Start() {
	serversNum := len(s.Servers)
	for port, c := range s.Servers {
		c.prepare()

		server := &http.Server{
			Addr:    fmt.Sprintf(":%v", port),
			Handler: c,
		}

		log.Printf("listening on %v", server.Addr)

		if serversNum > 1 {
			go func() {
				if err := server.ListenAndServe(); err != nil {
					log.Printf("E! ListenAndServe error: %v", err)
				}
			}()
		} else {
			if err := server.ListenAndServe(); err != nil {
				log.Printf("E! ListenAndServe error: %v", err)
			}
		}

		serversNum--
	}
}
