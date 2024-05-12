package middlewares

import (
	"net"
	"net/http"

	"go.uber.org/zap"
)

func TrustedSubnet(logger *zap.Logger, subnet *net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if subnet == nil {
				next.ServeHTTP(w, r)
				return
			}

			ipStr := r.Header.Get("X-Real-IP")
			if ipStr == "" {
				logger.Debug("X-Real-IP is empty")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			ip := net.ParseIP(ipStr)
			if ip == nil {
				logger.Error("Fail to parse X-Real-IP header", zap.String("ip", ipStr))
				w.WriteHeader(http.StatusForbidden)
				return
			}

			if !subnet.Contains(ip) {
				logger.Debug("X-Real-IP doesn't belong to the subnet", zap.String("ip", ipStr), zap.String("subnet", subnet.String()))
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
