package handler

import "net/http"

func (s *Server) HandleHey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Hey</title></head>
<body style="margin:0;display:flex;justify-content:center;align-items:center;min-height:100vh;background:#111">
<img src="https://api.thecatapi.com/v1/images/search?format=src" alt="Random cat" style="max-width:90vw;max-height:90vh">
</body>
</html>`))
}
