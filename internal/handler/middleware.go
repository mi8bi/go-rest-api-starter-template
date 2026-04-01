package handler

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

type contextKey string

const ctxUserID contextKey = "userID"

// Logger はリクエストのメソッド・パス・ステータス・処理時間をログに記録します。
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration", time.Since(start).String(),
		)
	})
}

// Recovery はpanicをキャッチして500を返します。
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered", "error", rec, "stack", string(debug.Stack()))
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Authenticate はCookieのセッショントークンを検証し、userIDをcontextにセットします。
// 未認証の場合は401を返します。
func Authenticate(sessions sessionGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err != nil {
				respondError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			userID, err := sessions.GetUserID(r.Context(), cookie.Value)
			if err != nil || userID == 0 {
				respondError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// userIDFromContext はcontextからuserIDを取り出します。
func userIDFromContext(ctx context.Context) int64 {
	id, _ := ctx.Value(ctxUserID).(int64)
	return id
}

// sessionGetter はAuthenticateミドルウェアが必要とする最小interface。
type sessionGetter interface {
	GetUserID(ctx context.Context, token string) (int64, error)
}

// responseWriter はステータスコードをキャプチャするためのラッパーです。
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
