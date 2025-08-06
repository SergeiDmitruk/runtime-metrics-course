package middleware

import (
	"net"
	"net/http"
)

type WhiteListMiddleware struct {
	subnet *net.IPNet
}

func NewWhiteListMiddleware(cidr string) *WhiteListMiddleware {
	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return &WhiteListMiddleware{subnet: nil}
	}
	return &WhiteListMiddleware{subnet: subnet}
}
func (h *WhiteListMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.subnet == nil {
			next.ServeHTTP(w, r)
			return
		}

		realIP := r.Header.Get("X-Real-IP")
		if realIP == "" {
			http.Error(w, "X-Real-IP header is required", http.StatusForbidden)
			return
		}
		ip := net.ParseIP(realIP)
		if ip == nil {
			http.Error(w, "Invalid IP format in X-Real-IP", http.StatusForbidden)
			return
		}

		if !h.subnet.Contains(ip) {
			http.Error(w, "IP not in whitelist", http.StatusForbidden)
			return
		} else {
			next.ServeHTTP(w, r)
		}

	})
}
