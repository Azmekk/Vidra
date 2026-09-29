package middleware

import (
	"net"
	"net/http"
	"net/netip"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

var privateNetworks = []string{
	"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "100.64.0.0/10",
	"::1/128", "fc00::/7", "fe80::/10",
}

// ClientIP resolves the client address. X-Forwarded-For is only honoured when
// the direct peer is on a private network, i.e. the reverse proxy in front of
// Vidra; public peers are keyed by their socket address and cannot spoof it.
func ClientIP(next http.Handler) http.Handler {
	prefixes := make([]netip.Prefix, len(privateNetworks))
	for i, p := range privateNetworks {
		prefixes[i] = netip.MustParsePrefix(p)
	}
	fromProxy := chimw.ClientIPFromXFF(privateNetworks...)(next)
	direct := chimw.ClientIPFromRemoteAddr(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Forwarded-For") != "" && isPrivate(r.RemoteAddr, prefixes) {
			fromProxy.ServeHTTP(w, r)
			return
		}
		direct.ServeHTTP(w, r)
	})
}

// RateLimitByIP limits requests per resolved client IP.
func RateLimitByIP(requests int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.LimitBy(requests, window, func(r *http.Request) (string, error) {
		return httprate.CanonicalizeIP(chimw.GetClientIP(r.Context())), nil
	})
}

func isPrivate(remoteAddr string, prefixes []netip.Prefix) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	ip = ip.Unmap().WithZone("")
	for _, p := range prefixes {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}
