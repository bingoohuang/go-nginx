package nginxconf

import (
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/bingoohuang/gonginx/directive"
)

func parseLocation(conf NginxConfigureCommand) (l directive.Location) {
	if len(conf.Words) == 2 {
		l.Path = conf.Words[1]
	} else {
		l.Modifier = directive.Modifier(conf.Words[1])
		l.Path = conf.Words[2]
	}

	if l.Priority = l.Modifier.Priority(); l.Priority == directive.ModifierRegular {
		reg := l.Path
		if l.Modifier == "~*" {
			reg = "(?i)" + reg
		}

		l.Pattern = regexp.MustCompile(reg)
	}

	l.Processors = make(directive.Processors, 0)

	for _, block := range conf.Block {
		directiveName := strings.ToLower(block.Words[0])

		if !l.Parse(directiveName, block.Words[1:]) {
			log.Printf("W! unsupported %+v", block.Words)
		}
	}

	sort.Sort(l.Processors)

	return l
}
