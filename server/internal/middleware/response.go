package middleware

import (
	"encoding/json"
	"net/http"
)

// writeJSON is a helper to write JSON responses from middleware.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
