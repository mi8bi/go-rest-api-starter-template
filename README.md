# go-rest-api-starter-template

Production-ready REST API template built with Go — 実務でそのまま使える最小構成のGoテンプレートです。

## 特徴

- **シンプルな構成** — `net/http` + SQLite のみ。過剰な抽象化なし
- **軽量クリーンアーキテクチャ** — handler / usecase / domain / repository / infra の5層
- **セッション認証** — CookieベースのセッションをSQLiteで管理
- **手動DI** — DIコンテナ不使用。`main.go` で明示的に組み立て
- **Graceful Shutdown** — SIGINT/SIGTERM で安全に停止
- **ミドルウェア** — ロギング・パニックリカバリ・認証チェック

## 必要環境

- Go 1.22+
- GCC（go-sqlite3のビルドに必要）

## セットアップ

```bash
git clone https://github.com/mi8bi/go-rest-api-starter-template.git
cd go-rest-api-starter-template

# 依存関係取得
go mod tidy

# 環境変数（任意）
cp .env.example .env
```

## 実行

```bash
# 開発環境
go run ./cmd/api

# ビルドして実行
go build -o bin/api ./cmd/api && ./bin/api

# 環境変数で設定変更
ADDR=:9090 DATABASE_DSN=./dev.db go run ./cmd/api
```

## API 一覧

### 認証不要

#### ユーザー登録

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Taro","email":"taro@example.com","password":"password123"}'
```

```json
{
  "id": 1,
  "name": "Taro",
  "email": "taro@example.com",
  "created_at": "2026-04-01T12:00:00Z"
}
```

#### ログイン

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"taro@example.com","password":"password123"}' \
  -c cookies.txt
```

Cookie に `session_token` がセットされます。

#### ログアウト

```bash
curl -X POST http://localhost:8080/logout -b cookies.txt -c cookies.txt
```

### 認証必須（Cookie が必要）

#### 自分の情報を取得

```bash
curl http://localhost:8080/me -b cookies.txt
```

#### ユーザー情報を取得

```bash
curl http://localhost:8080/users/1 -b cookies.txt
```

#### ユーザー名を更新（自分のみ）

```bash
curl -X PATCH http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Jiro"}' \
  -b cookies.txt
```

#### ユーザーを削除（自分のみ）

```bash
curl -X DELETE http://localhost:8080/users/1 -b cookies.txt
```

## エラーレスポンス

```json
{ "error": "invalid email or password" }
```

| ステータス | 意味 |
|---|---|
| 400 | バリデーションエラー |
| 401 | 未認証 |
| 403 | 権限なし |
| 404 | リソース未発見 |
| 409 | 重複（メールアドレス等） |
| 500 | サーバーエラー |

## ディレクトリ構成

```
.
├── cmd/
│   └── api/
│       └── main.go          # エントリポイント・手動DI・サーバー起動
├── internal/
│   ├── domain/
│   │   └── user.go          # エンティティ（純粋なGoの型）
│   ├── handler/
│   │   ├── auth_handler.go  # 登録・ログイン・ログアウト・me
│   │   ├── user_handler.go  # ユーザーCRUD
│   │   ├── middleware.go    # ログ・リカバリ・認証
│   │   └── response.go      # JSON/エラーヘルパー
│   ├── usecase/
│   │   ├── auth_usecase.go  # 認証ロジック
│   │   └── user_usecase.go  # ユーザーCRUDロジック
│   ├── repository/
│   │   └── user_repository.go  # UserRepository interface
│   └── infra/
│       ├── db.go                       # SQLite接続・マイグレーション
│       ├── sqlite_user_repository.go   # UserRepository実装
│       └── sqlite_session_store.go     # セッション実装
├── .env.example
├── go.mod
└── README.md
```

## アーキテクチャ

```
[HTTP] → handler → usecase → repository(interface)
                      ↓               ↑
                   domain          infra(実装)
```

- `domain` はGoの純粋な型。誰にも依存しない
- `repository` はinterfaceのみ。usecaseの隣に置く（使う側が定義）
- `infra` はrepository interfaceを実装するが、usecaseを知らない
- `main.go` だけがすべての層を知り、手動DIで組み立てる

## 設計思想

### なぜJWTではなくセッションなのか

- Cookieのみで完結するため実装がシンプル
- セッション無効化（ログアウト・強制失効）が即時に行える
- JWTは失効処理が複雑になりがちで、スターターには不向き

### なぜDIコンテナを使わないのか

- `main.go` を読めば依存関係が一目瞭然
- テンプレートとしての読みやすさを優先
- Go標準の手動DIで十分なスケールを扱える

### なぜSQLiteなのか

- セットアップ不要（ファイル1つ）
- 開発初期の素早いスタートに最適
- `UserRepository` interfaceを差し替えるだけでPostgreSQLに移行可能

## ライセンス

MIT
