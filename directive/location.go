package directive

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/bingoohuang/gonginx/util"
	"github.com/sirupsen/logrus"
)

type ProcessSeq int

const (
	Continue ProcessSeq = iota
	Terminate
)

type Processor interface {
	Naming
	GetProcessSeq() ProcessSeq
	Parse(path string, name string, params []string) error
	Do(l Location, w http.ResponseWriter, r *http.Request) ProcessResult
}

// nolint:gochecknoglobals
var factories = make([]ProcessorFactory, 0)

func RegisterFactory(n ProcessorFactory) {
	factories = append(factories, n)
}

type Processors []Processor

func (l Processors) Len() int { return len(l) }

func (l Processors) Less(i, j int) bool {
	return l[i].GetProcessSeq() < l[j].GetProcessSeq()
}

func (l Processors) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

type ProcessResult int

const (
	ProcessContinue ProcessResult = iota
	ProcessTerminate
)

type Naming interface {
	Name() map[string]bool
}

type ProcessorFactory interface {
	Naming
	Create() Processor
}

// Location is the location.
type Location struct {
	Seq        int              // 定义的顺序
	Priority   ModifierPriority // 匹配级别，从0开始，数字越小，匹配优先级越高
	Modifier   Modifier
	Path       string
	Processors Processors
	Pattern    *regexp.Regexp
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

func (l Location) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, v := range l.Processors {
		if v.Do(l, w, r) == ProcessTerminate {
			break
		}
	}
}

func (l *Location) createProcessor(firstWord string) Processor {
	for _, v := range factories {
		if v.Name()[firstWord] {
			return v.Create()
		}
	}

	return nil
}

func (l *Location) findProcessor(firstWord string) Processor {
	for _, v := range l.Processors {
		if v.Name()[firstWord] {
			return v
		}
	}

	return nil
}

func (l *Location) Parse(directive string, params []string) bool {
	dp := l.findProcessor(directive)
	if dp == nil {
		dp = l.createProcessor(directive)
		l.Processors = append(l.Processors, dp)
	}

	if err := dp.Parse(l.Path, directive, params); err != nil {
		logrus.Fatalf("invalid conf for %v error %+v", params, err)
		return false
	}

	return true
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
