package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"victortillett.net/test1-national-inservice-api/internal/models"
	"victortillett.net/test1-national-inservice-api/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// auth request/response DTOs
type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"`
	PersonID *int64 `json:"person_id,omitempty"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type AuthHandler struct {
	db     *pgxpool.Pool
	jwtKey []byte
	email  services.Emailer
	// token expiry in hours
	expHours int
}

func NewAuthHandler(db *pgxpool.Pool, jwtKey []byte, expHours int, emailer services.Emailer) *AuthHandler {
	return &AuthHandler{db: db, jwtKey: jwtKey, email: emailer, expHours: expHours}
}

// helper: hash password
func hashPassword(p string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
}

func comparePassword(hash []byte, password string) error {
	return bcrypt.CompareHashAndPassword(hash, []byte(password))
}

// Register: create account with hashed password
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	pwHash, err := hashPassword(req.Password)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	acc := &models.Account{
		PersonID:     req.PersonID,
		Email:        req.Email,
		PasswordHash: string(pwHash),
		Role:         req.Role,
		IsActive:     true, // or false if you want activation flow
		CreatedAt:    time.Now(),
	}

	id, err := models.CreateAccount(ctx, tx, acc)
	if err != nil {
		http.Error(w, "insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "commit error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"account_id": id})
}

// Login: verify credentials and return JWT
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	acc, err := models.GetAccountByEmail(ctx, h.db, req.Email)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if acc == nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := comparePassword([]byte(acc.PasswordHash), req.Password); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Build JWT claims
	claims := jwt.MapClaims{
		"sub":  acc.AccountID,
		"role": acc.Role,
		"exp":  time.Now().Add(time.Duration(h.expHours) * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(h.jwtKey)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(loginResponse{Token: signed})
}

// ForgotPassword: generate token and send email (simple stub)
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Email string `json:"email"`
	}
	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	acc, err := models.GetAccountByEmail(ctx, h.db, body.Email)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if acc == nil {
		// don't reveal whether email exists
		w.WriteHeader(http.StatusAccepted)
		return
	}

	// create a simple one-time token (signed JWT short lived)
	claims := jwt.MapClaims{
		"sub": acc.AccountID,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"op":  "pwreset",
	}
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tkn.SignedString(h.jwtKey)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	// send email (async - fire and forget)
	go func() {
		ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		subject := "Password reset"
		body := "Use this token to reset your password: " + signed
		_ = h.email.Send(ctx2, acc.Email, subject, body) // ignore error, but you may log it
	}()

	w.WriteHeader(http.StatusAccepted)
}

// ResetPassword: accept token + new password
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Token == "" || body.Password == "" {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	// parse token
	tkn, err := jwt.Parse(body.Token, func(t *jwt.Token) (interface{}, error) {
		// only support HS256
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("invalid token")
		}
		return h.jwtKey, nil
	})
	if err != nil || !tkn.Valid {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}
	claims, ok := tkn.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}
	// ensure op == pwreset
	if op, _ := claims["op"].(string); op != "pwreset" {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}
	sub, ok := claims["sub"].(float64)
	if !ok {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}
	accID := int64(sub)

	// hash new password
	pwHash, err := hashPassword(body.Password)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// load account (optional)
	acc, err := models.GetAccountByID(ctx, h.db, accID)
	if err != nil || acc == nil {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	acc.PasswordHash = string(pwHash)
	if err := models.UpdateAccount(ctx, tx, acc); err != nil {
		http.Error(w, "update error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "commit error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
