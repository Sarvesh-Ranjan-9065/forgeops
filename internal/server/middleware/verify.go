// Package middleware provides HTTP middleware for the ForgeOps webhook server.
package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
)

const signatureHeader = "X-Hub-Signature-256"

// VerifySignature returns middleware that verifies the GitHub HMAC-SHA256
// signature of each request body using secret. Requests with a missing or
// invalid signature are rejected with 401. The body is restored so downstream
// handlers can read it.
func VerifySignature(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "cannot read body", http.StatusBadRequest)
				return
			}
			_ = r.Body.Close()
			r.Body = io.NopCloser(bytes.NewReader(body))

			if !validSignature(secret, body, r.Header.Get(signatureHeader)) {
				http.Error(w, "invalid signature", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// validSignature reports whether got matches the expected HMAC-SHA256 signature
// of body in the form "sha256=<hex>". The comparison is constant time.
func validSignature(secret, body []byte, got string) bool {
	const prefix = "sha256="
	if !strings.HasPrefix(got, prefix) {
		return false
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	want := prefix + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(want), []byte(got))
}
