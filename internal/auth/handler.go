package auth

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	repo      *Repository
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewHandler(repo *Repository, jwtSecret string, tokenTTL time.Duration) *Handler {
	if tokenTTL <= 0 {
		tokenTTL = 15 * time.Minute
	}
	return &Handler{repo: repo, jwtSecret: []byte(jwtSecret), tokenTTL: tokenTTL}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var input RegisterRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Name == "" || utf8.RuneCountInString(input.Name) > 100 {
		writeError(w, http.StatusBadRequest, "name is required and must be at most 100 characters")
		return
	}
	if len(input.Email) > 255 || !validEmail(input.Email) {
		writeError(w, http.StatusBadRequest, "valid email is required")
		return
	}
	// bcrypt only accepts passwords up to 72 bytes.
	if len(input.Password) < 8 || len(input.Password) > 72 {
		writeError(w, http.StatusBadRequest, "password must be between 8 and 72 bytes")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("hash password: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	u, err := h.repo.Create(r.Context(), input.Name, input.Email, string(hash))
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		log.Printf("register user: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response, err := h.newAuthResponse(u)
	if err != nil {
		log.Printf("create access token: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var input LoginRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if !validEmail(input.Email) || input.Password == "" || len(input.Password) > 72 {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	u, err := h.repo.GetByEmail(r.Context(), input.Email)
	if err != nil {
		if !errors.Is(err, ErrUserNotFound) {
			log.Printf("find login user: %v", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(input.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	response, err := h.newAuthResponse(u)
	if err != nil {
		log.Printf("create access token: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	id, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	u, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			writeError(w, http.StatusUnauthorized, "user no longer exists")
			return
		}
		log.Printf("get current user: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(u))
}

func (h *Handler) newAuthResponse(u User) (AuthResponse, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   u.ID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(h.tokenTTL)),
		Issuer:    "forgeflow",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(h.jwtSecret)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.tokenTTL.Seconds()),
		User:        toUserResponse(u),
	}, nil
}

func toUserResponse(u User) UserResponse {
	return UserResponse{ID: u.ID.String(), Name: u.Name, Email: u.Email, CreatedAt: u.CreatedAt}
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("request must contain exactly one JSON value")
	}
	return nil
}

func validEmail(value string) bool {
	addr, err := mail.ParseAddress(value)
	return err == nil && addr.Address == value && strings.Contains(value, "@")
}
