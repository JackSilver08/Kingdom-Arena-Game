package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Elo      int    `json:"elo"`
	Wins     int    `json:"wins"`
	Losses   int    `json:"losses"`
}

type Handler struct {
	jwtSecret string
	users     map[string]*user
}

func NewHandler(jwtSecret string) *Handler {
	return &Handler{
		jwtSecret: jwtSecret,
		users:     make(map[string]*user),
	}
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if _, exists := h.users[req.Email]; exists {
		http.Error(w, `{"error":"email already registered"}`, http.StatusConflict)
		return
	}

	hashedPw, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	u := &user{
		ID:       "user_" + req.Username,
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPw),
		Elo:      1000,
	}
	h.users[req.Email] = u

	token, _ := h.generateToken(u)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse{Token: token})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	u, exists := h.users[req.Email]
	if !exists {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token, _ := h.generateToken(u)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse{Token: token})
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	for _, u := range h.users {
		if u.ID == userID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(u)
			return
		}
	}
	http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
}

func (h *Handler) generateToken(u *user) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  u.ID,
		"username": u.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}
