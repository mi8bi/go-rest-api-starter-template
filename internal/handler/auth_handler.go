package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/mi8bi/go-rest-api-starter-template/internal/domain"
	"github.com/mi8bi/go-rest-api-starter-template/internal/usecase"
)

// authService はAuthHandlerが必要とするusecase操作のinterface。
type authService interface {
	Register(ctx context.Context, name, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	Logout(ctx context.Context, token string) error
	Me(ctx context.Context, token string) (*domain.User, error)
}

// AuthHandler は認証系エンドポイントを処理します。
type AuthHandler struct {
	auth authService
}

func NewAuthHandler(auth authService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// POST /register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "name, email, password are required")
		return
	}
	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	u, err := h.auth.Register(r.Context(), req.Name, req.Email, req.Password)
	if errors.Is(err, usecase.ErrEmailAlreadyExists) {
		respondError(w, http.StatusConflict, "email already exists")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusCreated, toUserResponse(u))
}

// POST /login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if errors.Is(err, usecase.ErrInvalidCredentials) {
		respondError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,                       // XSS対策：JSからアクセス不可
		Secure:   false,                      // 本番ではtrueにすること（HTTPS必須）
		SameSite: http.SameSiteLaxMode,       // CSRF対策
		Expires:  time.Now().Add(24 * time.Hour),
	})

	respondJSON(w, http.StatusOK, map[string]string{"message": "logged in"})
}

// POST /logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		respondJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
		return
	}

	_ = h.auth.Logout(r.Context(), cookie.Value)

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Path:    "/",
		MaxAge:  -1, // Cookieを即削除
		HttpOnly: true,
	})

	respondJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// GET /me  ← 認証必須
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	cookie, err := r.Cookie("session_token")
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	u, err := h.auth.Me(r.Context(), cookie.Value)
	if err != nil || u == nil || u.ID != userID {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	respondJSON(w, http.StatusOK, toUserResponse(u))
}

// userResponse はレスポンス用の安全なユーザー表現（パスワードハッシュを除外）。
type userResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}
