package linkhub

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/logback"
)

type Server struct {
	log logback.Logger
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//if r.Method != http.MethodConnect {
	//}
	//
	//hijacker, ok := w.(http.Hijacker)
	//if !ok {
	//	return
	//}
	//conn, rw, err := hijacker.Hijack()
	//if err != nil {
	//	return
	//}
}
