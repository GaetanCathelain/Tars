package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/GaetanCathelain/Tars/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	DB        *pgxpool.Pool
	JWTSecret string
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, "username and password required", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		slog.Error("hash password", "error", err)
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	var userID string
	err = s.DB.QueryRow(r.Context(),
		"INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id",
		req.Username, hash,
	).Scan(&userID)
	if err != nil {
		slog.Error("create user", "error", err)
		writeError(w, "username already taken", http.StatusConflict)
		return
	}

	token, err := auth.GenerateToken(s.JWTSecret, mustParseUUID(userID), req.Username)
	if err != nil {
		slog.Error("generate token", "error", err)
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(authResponse{Token: token})
}

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, "username and password required", http.StatusBadRequest)
		return
	}

	var userID, hash string
	err := s.DB.QueryRow(r.Context(),
		"SELECT id, password FROM users WHERE username = $1",
		req.Username,
	).Scan(&userID, &hash)
	if err != nil {
		writeError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := auth.CheckPassword(hash, req.Password); err != nil {
		writeError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(s.JWTSecret, mustParseUUID(userID), req.Username)
	if err != nil {
		slog.Error("generate token", "error", err)
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse{Token: token})
}

func writeError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
