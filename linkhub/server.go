package linkhub

import (
	"net/http"
)

type Server struct{}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me")
}
