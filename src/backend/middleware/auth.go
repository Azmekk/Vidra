package middleware

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Azmekk/Vidra/backend/services"
	"github.com/Azmekk/Vidra/backend/utils"
)

const SessionCookie = "vidra_session"

type ctxKey int

const (
	userKey ctxKey = iota
	tokenKey
	viaKey
)

// UserFrom returns the authenticated user stored by RequireAuth.
func UserFrom(ctx context.Context) (services.User, bool) {
	u, ok := ctx.Value(userKey).(services.User)
	return u, ok
}

// AuthMethod reports how the request was authenticated: "session" or "token".
func AuthMethod(ctx context.Context) string {
	m, _ := ctx.Value(viaKey).(string)
	return m
}

// SessionTokenFrom returns the session token used for the request, if any.
func SessionTokenFrom(ctx context.Context) string {
	t, _ := ctx.Value(tokenKey).(string)
	return t
}

type Auth struct {
	Service         *services.AuthService
	InsecureCookies bool
}

// RequireAuth accepts a session cookie or an API bearer token. Cookie
// authenticated writes must come from the same origin.
func (a *Auth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		scheme, token, _ := strings.Cut(r.Header.Get("Authorization"), " ")
		if strings.EqualFold(scheme, "Bearer") {
			user, err := a.Service.AuthenticateAPIToken(ctx, strings.TrimSpace(token))
			if err != nil {
				utils.Annotate(w, "reason", err.Error())
				utils.RespondWithError(w, http.StatusUnauthorized, "invalid API token")
				return
			}
			ctx = context.WithValue(ctx, userKey, user)
			ctx = context.WithValue(ctx, viaKey, "token")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		cookie, err := r.Cookie(SessionCookie)
		if err != nil {
			reason := "no session cookie or API token"
			if scheme != "" {
				reason = "Authorization header must be 'Bearer <token>'"
			}
			utils.Annotate(w, "reason", reason)
			utils.RespondWithError(w, http.StatusUnauthorized, "not authenticated")
			return
		}
		user, refreshed, err := a.Service.Authenticate(ctx, cookie.Value)
		if err != nil {
			a.ClearCookie(w)
			utils.RespondWithError(w, http.StatusUnauthorized, "session expired")
			return
		}
		if !isSafeMethod(r.Method) && !sameOrigin(r) {
			utils.Annotate(w, "origin", r.Header.Get("Origin"), "host", r.Host)
			utils.RespondWithError(w, http.StatusForbidden, "cross-origin request blocked")
			return
		}
		if refreshed != nil {
			a.SetCookie(w, *refreshed)
		}
		ctx = context.WithValue(ctx, userKey, user)
		ctx = context.WithValue(ctx, tokenKey, cookie.Value)
		ctx = context.WithValue(ctx, viaKey, "session")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Auth) SetCookie(w http.ResponseWriter, s services.Session) {
	c := &http.Cookie{
		Name:     SessionCookie,
		Value:    s.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   !a.InsecureCookies,
		SameSite: http.SameSiteLaxMode,
	}
	if s.Remember {
		c.Expires = s.ExpiresAt
		c.MaxAge = int(time.Until(s.ExpiresAt).Seconds())
	}
	http.SetCookie(w, c)
}

func (a *Auth) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: SessionCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: !a.InsecureCookies, SameSite: http.SameSiteLaxMode,
	})
}

func isSafeMethod(m string) bool {
	return m == http.MethodGet || m == http.MethodHead || m == http.MethodOptions
}

func sameOrigin(r *http.Request) bool {
	switch r.Header.Get("Sec-Fetch-Site") {
	case "same-origin", "none":
		return true
	case "":
	default:
		return false
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	return err == nil && u.Host == r.Host
}
