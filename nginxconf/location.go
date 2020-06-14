package nginxconf

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/bingoohuang/gonginx/util"
)

type ProcessResult int

const (
	ProcessContinue ProcessResult = iota
	ProcessTerminate
)

type Naming interface {
	Name() map[string]bool
}

type LocationProcessorFactory interface {
	Naming
	Create() LocationProcessor
}

type ProcessSeq int

const (
	Terminate ProcessSeq = iota
	Continue
)

type LocationProcessor interface {
	Naming
	GetProcessSeq() ProcessSeq
	Parse(path string, name string, params []string) error
	Do(l Location, w http.ResponseWriter, r *http.Request) ProcessResult
}

// nolint gochecknoglobals
var LocationFactories = make([]LocationProcessorFactory, 0)

func AppendLocationFactory(n LocationProcessorFactory) {
	LocationFactories = append(LocationFactories, n)
}

type LocationProcessors []LocationProcessor

func (l LocationProcessors) Len() int { return len(l) }

func (l LocationProcessors) Less(i, j int) bool {
	return l[i].GetProcessSeq() < l[j].GetProcessSeq()
}

func (l LocationProcessors) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// Location is the location.
type Location struct {
	Seq        int              // 定义的顺序
	Priority   ModifierPriority // 匹配级别，从0开始，数字越小，匹配优先级越高
	Modifier   Modifier
	Path       string
	Processors LocationProcessors
	//Index       string
	//Root        string
	//Alias string
	//ProxyPass   *ProxyPass
	//Echo        *Echo
	//Return      *Return
	Pattern *regexp.Regexp
	//DefaultType *DefaultType
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
