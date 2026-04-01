package handler

import (
	"encoding/json"
	"net/http"
)

// respondJSON はJSONレスポンスを書き込みます。
func respondJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// respondError はエラーレスポンスを統一フォーマットで返します。
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
