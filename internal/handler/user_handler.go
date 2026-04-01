package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/mi8bi/go-rest-api-starter-template/internal/domain"
	"github.com/mi8bi/go-rest-api-starter-template/internal/usecase"
)

// userService はUserHandlerが必要とするusecase操作のinterface。
type userService interface {
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	UpdateName(ctx context.Context, id int64, name string) (*domain.User, error)
	Delete(ctx context.Context, id int64) error
}

// UserHandler はユーザーCRUDエンドポイントを処理します。
type UserHandler struct {
	users userService
}

func NewUserHandler(users userService) *UserHandler {
	return &UserHandler{users: users}
}

// GET /users/{id}
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}

	u, err := h.users.GetByID(r.Context(), id)
	if errors.Is(err, usecase.ErrUserNotFound) {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, toUserResponse(u))
}

// PATCH /users/{id}
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}

	// 自分自身のみ更新可（認証済みuserIDと照合）
	if userIDFromContext(r.Context()) != id {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	u, err := h.users.UpdateName(r.Context(), id, req.Name)
	if errors.Is(err, usecase.ErrUserNotFound) {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, toUserResponse(u))
}

// DELETE /users/{id}
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}

	if userIDFromContext(r.Context()) != id {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := h.users.Delete(r.Context(), id); errors.Is(err, usecase.ErrUserNotFound) {
		respondError(w, http.StatusNotFound, "user not found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// parseIDFromPath は /users/123 形式のパスからIDを取り出します。
// Go 1.22+ の net/http はパスパラメータをサポートしますが、
// 本テンプレートではシンプルに strings.Split を使います。
func parseIDFromPath(w http.ResponseWriter, path string) (int64, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		respondError(w, http.StatusBadRequest, "invalid path")
		return 0, false
	}
	id, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}
