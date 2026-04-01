package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mi8bi/go-rest-api-starter-template/internal/handler"
	"github.com/mi8bi/go-rest-api-starter-template/internal/infra"
	"github.com/mi8bi/go-rest-api-starter-template/internal/repository"
	"github.com/mi8bi/go-rest-api-starter-template/internal/usecase"
)

func main() {
	// ── 設定 ──────────────────────────────────────────────
	addr := getEnv("ADDR", ":8080")
	dsn := getEnv("DATABASE_DSN", "./app.db")

	// ── インフラ初期化 ────────────────────────────────────
	db, err := infra.NewDB(dsn)
	if err != nil {
		slog.Error("failed to open db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// ── 手動DI（DIコンテナは使わない） ───────────────────
	userRepo := repository.NewSQLiteUserRepository(db)
	sessionStore := infra.NewSQLiteSessionStore(db)

	authUC := usecase.NewAuthUsecase(userRepo, sessionStore)
	userUC := usecase.NewUserUsecase(userRepo)

	authH := handler.NewAuthHandler(authUC)
	userH := handler.NewUserHandler(userUC)

	// ── ルーティング ──────────────────────────────────────
	// Go 1.22+ では "METHOD /path" 形式でメソッドを指定できます。
	mux := http.NewServeMux()

	// 公開エンドポイント
	mux.HandleFunc("POST /register", authH.Register)
	mux.HandleFunc("POST /login", authH.Login)
	mux.HandleFunc("POST /logout", authH.Logout)

	// 認証必須エンドポイント
	auth := handler.Authenticate(sessionStore)
	mux.Handle("GET /me", auth(http.HandlerFunc(authH.Me)))
	mux.Handle("GET /users/{id}", auth(http.HandlerFunc(userH.GetByID)))
	mux.Handle("PATCH /users/{id}", auth(http.HandlerFunc(userH.Update)))
	mux.Handle("DELETE /users/{id}", auth(http.HandlerFunc(userH.Delete)))

	// ── グローバルミドルウェア適用 ────────────────────────
	h := handler.Logger(handler.Recovery(mux))

	// ── サーバー起動 ──────────────────────────────────────
	srv := &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("server stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
