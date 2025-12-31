package middleware

import (
	"net/http"
)

// SecurityMiddleware adds security headers to responses
func SecurityMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent clickjacking attacks
			w.Header().Set("X-Frame-Options", "DENY")

			// Prevent MIME type sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Enable browser XSS protection
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Strict Transport Security (HSTS)
			// max-age=31536000 (1 year), includeSubDomains
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

			// Content Security Policy
			// Adjust based on your application needs
			csp := "default-src 'self'; " +
				"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
				"style-src 'self' 'unsafe-inline'; " +
				"img-src 'self' data: https:; " +
				"font-src 'self' data:; " +
				"connect-src 'self'; " +
				"frame-ancestors 'none'"
			w.Header().Set("Content-Security-Policy", csp)

			// Referrer Policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions Policy (formerly Feature Policy)
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			next.ServeHTTP(w, r)
		})
	}
}
