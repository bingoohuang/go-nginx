package directive

import "net/http"

// nolint:gochecknoinits
func init() {
	RegisterFactory(&defaultTypeNaming{})
}

type defaultTypeNaming struct{}

func (i defaultTypeNaming) Create() Processor {
	return &defaultType{defaultTypeNaming: i}
}

func (defaultTypeNaming) Name() map[string]bool {
	return map[string]bool{
		"default_type": true,
	}
}

type defaultType struct {
	defaultTypeNaming

	Value string
}

func (r *defaultType) GetProcessSeq() ProcessSeq { return Continue }

func (r *defaultType) Do(l Location, w http.ResponseWriter, rq *http.Request) ProcessResult {
	w.Header().Set("Content-Type", r.Value)
	return ProcessContinue
}

func (r *defaultType) Parse(path string, name string, params []string) error {
	r.Value = params[0]
	return nil
}
