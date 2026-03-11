package handler

import (
	"net/http"
)

func (s *Server) HandleCreateWorker(w http.ResponseWriter, r *http.Request) {
	writeError(w, "not implemented", http.StatusNotImplemented)
}
