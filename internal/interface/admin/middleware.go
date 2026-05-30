package admin

import (
	"net/http"
	"strings"
)

// AdminAuthMiddleware 校验 admin_api_key
func AdminAuthMiddleware(adminAPIKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "缺少认证信息"})
				return
			}

			token := auth
			if strings.HasPrefix(auth, "Bearer ") {
				token = strings.TrimSpace(auth[7:])
			}

			if token != adminAPIKey {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "API Key 无效"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
