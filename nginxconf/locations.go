package nginxconf

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/bingoohuang/gonginx/util"
	"github.com/sirupsen/logrus"
)

type Locations []Location

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
			l.ProxyPass = ProxyPassParse(l.Path, block.Words[1:])
		case "echo":
			l.Echo = EchoParse(l.Path, block.Words[1:])
		case "return":
			l.Return = ReturnParse(block.Words[1:])
		case "default_type":
			l.DefaultType = DefaultTypeParse(block.Words[1:])
		default:
			logrus.Warnf("unsupported %+v", block.Words)
		}
	}

	return l
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
