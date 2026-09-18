// Package middleware provides HTTP middleware for the ForgeOps webhook server.
package middleware

import (
	"crypto/subtle"
	"net"
	"net/http"
	"strings"
)

// ProtectInternal returns middleware that restricts access to internal endpoints
// such as /metrics and /dashboard. If username and password are provided, valid
// HTTP Basic Authentication is required. If credentials are not configured, only
// requests originating from loopback (127.0.0.1, ::1) or private RFC1918/RFC4193
// networks are allowed; public internet requests are rejected with 403 Forbidden.
func ProtectInternal(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if username != "" && password != "" {
				user, pass, ok := r.BasicAuth()
				if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 ||
					subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
					w.Header().Set("WWW-Authenticate", `Basic realm="ForgeOps Admin"`)
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// If no basic auth credentials are configured, restrict to private/loopback callers.
			if isPrivateOrLoopback(r.RemoteAddr) {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, "forbidden: private endpoint", http.StatusForbidden)
		})
	}
}

// isPrivateOrLoopback reports whether the remote address represents a loopback,
// private network, or test network address.
func isPrivateOrLoopback(remoteAddr string) bool {
	if remoteAddr == "" {
		return true
	}
	host := remoteAddr
	if h, _, err := net.SplitHostPort(remoteAddr); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	if idx := strings.Index(host, "%"); idx != -1 {
		host = host[:idx]
	}
	if host == "localhost" || host == "" {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || host == "192.0.2.1"
}
