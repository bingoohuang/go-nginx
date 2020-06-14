package nginxconf

import (
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/gonginx/util"
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

	l.Processors = make([]LocationProcessor, 0)

	for _, block := range conf.Block {
		firstWord := strings.ToLower(block.Words[0])

		if l.existsProcessorParse(firstWord, block) {
			continue
		}

		if l.createProcessor(firstWord, block) {
			continue
		}

		logrus.Warnf("unsupported %+v", block.Words)
	}

	sort.Sort(l.Processors)

	return l
}

func (l *Location) createProcessor(firstWord string, block NginxConfigureCommand) bool {
	for _, v := range LocationFactories {
		if v.Name()[firstWord] {
			p := v.Create()
			if err := p.Parse(l.Path, firstWord, block.Words[1:]); err != nil {
				logrus.Fatalf("invalid conf for %v error %+v", block, err)
			}

			l.Processors = append(l.Processors, p)

			return true
		}
	}

	return false
}

func (l *Location) existsProcessorParse(firstWord string, block NginxConfigureCommand) bool {
	for _, v := range l.Processors {
		if v.Name()[firstWord] {
			if err := v.Parse(l.Path, firstWord, block.Words[1:]); err != nil {
				logrus.Fatalf("invalid conf for %v error %+v", block, err)
			}

			return true
		}
	}

	return false
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
