package directive

import (
	"fmt"
	"net/http"

	"github.com/bingoohuang/gou/str"
)

func init() {
	RegisterFactory(&returnNaming{})
}

type returnNaming struct{}

func (i returnNaming) Create() Processor {
	return &Return{returnNaming: i}
}

func (returnNaming) Name() map[string]bool {
	return map[string]bool{
		"return": true,
	}
}

// Return means http://nginx.org/en/docs/http/ngx_http_rewrite_module.html#return.
// Syntax:	return code [text];.
type Return struct {
	returnNaming

	Code int
	Text string
}

func (r *Return) GetProcessSeq() ProcessSeq {
	return Terminate
}

func (r *Return) Do(l Location, w http.ResponseWriter, rq *http.Request) ProcessResult {
	w.WriteHeader(r.Code)

	if r.Text != "" {
		_, _ = fmt.Fprint(w, r.Text)
	}

	return ProcessContinue
}

func (r *Return) Parse(path string, name string, params []string) error {
	r.Code = str.ParseInt(params[0])

	if len(params) > 1 {
		r.Text = params[1]
	}

	return nil
}
