package nginxconf

import (
	"fmt"
	"net/http"
	"strings"
)

// nolint gochecknoinits
func init() {
	AppendLocationFactory(&echoNaming{})
}

type echoNaming struct{}

func (i echoNaming) Create() LocationProcessor {
	return &echo{echoNaming: i, Values: make([]string, 0)}
}

func (echoNaming) Name() map[string]bool {
	return map[string]bool{
		"echo": true,
	}
}

type echo struct {
	echoNaming

	Values []string
}

func (r *echo) Parse(path string, name string, params []string) error {
	r.Values = append(r.Values, params...)

	return nil
}

func (r *echo) GetProcessSeq() ProcessSeq { return Continue }

func (r *echo) Do(l Location, w http.ResponseWriter, rq *http.Request) ProcessResult {
	for _, v := range r.Values {
		v = strings.ReplaceAll(v, "$request", rq.RequestURI)
		_, _ = fmt.Fprintln(w, v)
	}

	return ProcessContinue
}
