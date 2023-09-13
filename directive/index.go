package directive

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/bingoohuang/gou/file"
)

func init() {
	RegisterFactory(&indexNaming{})
}

type indexNaming struct{}

func (i indexNaming) Create() Processor {
	return &index{indexNaming: i}
}

func (indexNaming) Name() map[string]bool {
	return map[string]bool{
		"index": true,
		"root":  true,
		"alias": true,
	}
}

type index struct {
	indexNaming

	Index string
	Root  string
	Alias string
}

func (i *index) GetProcessSeq() ProcessSeq { return Terminate }

// ErrSyntax means that a syntax error occurred.

var ErrSyntax = errors.New("syntax error")

func (i *index) Parse(path string, name string, params []string) error {
	if len(params) == 0 {
		return ErrSyntax
	}

	switch name {
	case "index":
		i.Index = params[0]
	case "root":
		i.Root = params[0]
	case "alias":
		i.Alias = params[0]
	}

	return nil
}

func (i *index) Do(l Location, w http.ResponseWriter, r *http.Request) ProcessResult {
	// http://nginx.org/en/docs/http/ngx_http_index_module.html
	// processes requests ending with the slash character (‘/’).
	if strings.HasSuffix(r.URL.Path, "/") {
		http.Redirect(w, r, filepath.Join(r.URL.Path, i.Index), http.StatusFound)

		return ProcessTerminate
	}

	serveFile := r.URL.Path

	switch {
	case i.Root != "":
		// http://nginx.org/en/docs/http/ngx_http_core_module.html#root
		serveFile = filepath.Join(i.Root, serveFile)
	case i.Alias != "":
		// http://nginx.org/en/docs/http/ngx_http_core_module.html#alias
		// location /i/ { alias /data/w3/images/; }
		// on request of “/i/top.gif”, the file /data/w3/images/top.gif will be sent.
		serveFile = filepath.Join(i.Alias, strings.TrimPrefix(r.URL.Path, l.Path))
	default:
		serveFile = strings.TrimPrefix(serveFile, "/")
	}

	if r.URL.Path == "/" && file.SingleFileExists(serveFile) != nil {
		return Welcome(w)
	}

	if strings.HasSuffix(serveFile, "/index.html") {
		r.URL.Path = "avoid index.html redirect... in ServeHTTP"
	}

	http.ServeFile(w, r, serveFile)

	return ProcessTerminate
}
