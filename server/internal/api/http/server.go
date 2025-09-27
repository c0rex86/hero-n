package httpapi

import (
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	addr string
}

func New(addr string) *Server { return &Server{addr: addr} }

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})
	log.Printf("http listening on %s", s.addr)
	return http.ListenAndServe(s.addr, mux)
}
