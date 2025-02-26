package prof

import (
	"net/http"
	"path"

	"github.com/google/pprof/driver"
)

func New(name string) (http.Handler, error) {
	ph := new(pprofHandler)
	set := &pprofFlag{Args: []string{"-http=localhost:0", "-no_browser", name}}
	opt := &driver.Options{
		Flagset:    set,
		HTTPServer: ph.profHTTP,
	}
	if err := driver.PProf(opt); err != nil {
		return nil, err
	}

	return ph, nil
}

type pprofHandler struct {
	args *driver.HTTPServerArgs
}

func (ph *pprofHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := ph.matchHandler(r.URL.Path)
	if h == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	h.ServeHTTP(w, r)
}

func (ph *pprofHandler) profHTTP(args *driver.HTTPServerArgs) error {
	ph.args = args
	return nil
}

func (ph *pprofHandler) matchHandler(pth string) http.Handler {
	args := ph.args
	if args == nil {
		return nil
	}
	handlers := args.Handlers
	if handlers == nil {
		return nil
	}

	uri := path.Clean(pth)
	h := handlers[uri]

	return h
}
