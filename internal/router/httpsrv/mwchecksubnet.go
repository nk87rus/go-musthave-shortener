package httpsrv

import (
	"net/http"
	"net/netip"
	"strings"
)

func (s *Server) checkSubnetMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.trustedPrefix == nil {
			http.Error(w, "Доступ запрещён: не указан префикс доверенной сети", http.StatusForbidden)
			return
		}

		const realIPHeader = "X-Real-IP"
		realIP := r.Header.Get(realIPHeader)
		if realIP == "" {
			http.Error(w, "Доступ запрещён: не найден заголовок "+realIPHeader, http.StatusForbidden)
			return
		}

		ipStr := strings.TrimSpace(strings.Split(realIP, ",")[0])
		ip, err := netip.ParseAddr(ipStr)
		if err != nil {
			http.Error(w, "Доступ запрещён: ошибка при анализе адреса, указанного в заголовке "+realIPHeader+": "+err.Error(), http.StatusForbidden)
			return
		}

		if !s.trustedPrefix.Contains(ip) {
			http.Error(w, "Доступ запрещён: IP "+ip.String()+" не соответсвует префиксу доверенной сети ", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
