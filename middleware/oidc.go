package middleware

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/stuart-warren/serveit/oidc"
)

type Verifier interface {
	Verify(ctx context.Context, redirectTo string) (*oidc.IDToken, error)
}

type OIDCMiddlewareConfig struct {
	oidcAuth     Verifier
	redirectTo   string
	dontRedirect []string
	secureCookie bool
	now          func() time.Time
}

func NewOIDCMiddlewareConfig(oidcAuth Verifier) OIDCMiddlewareConfig {
	return OIDCMiddlewareConfig{
		oidcAuth:     oidcAuth,
		now:          time.Now,
		secureCookie: true,
		dontRedirect: []string{"/auth", "/callback"},
		redirectTo:   "/auth",
	}
}

func OIDC(c OIDCMiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, cookie := range r.Cookies() {
				if cookie.Name == oidc.JWTCookie {
					jwt, err := base64.URLEncoding.DecodeString(cookie.Value)
					if err != nil {
						http.Error(w, "could not decode cookie", http.StatusBadRequest)
						return
					}
					tkn, err := c.oidcAuth.Verify(r.Context(), string(jwt))
					if err != nil {
						log.Printf("invalid jwt: %s", err)
						break
					}

					var claims struct {
						Email         string `json:"email"`
						EmailVerified bool   `json:"email_verified"`
					}
					err = tkn.Claims(&claims)
					if claims.EmailVerified {
						r.Header.Set("User", claims.Email)
					}
					r.Header.Set("Authorisation", fmt.Sprintf("Bearer %s", jwt))
					w.Header().Set(oidc.JWTHeader, string(jwt))
					next.ServeHTTP(w, r)
					return

				}
			}
			for _, urlSuffix := range c.dontRedirect {
				if strings.HasSuffix(r.URL.Path, urlSuffix) {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.SetCookie(w, &http.Cookie{
				Name:     oidc.RedirectCookie,
				Value:    base64.URLEncoding.EncodeToString([]byte(r.RequestURI)),
				Expires:  c.now().Add(1 * time.Minute),
				Secure:   c.secureCookie,
				SameSite: http.SameSiteStrictMode,
				Path:     "/",
				HttpOnly: true,
			})
			http.Redirect(w, r, c.redirectTo, http.StatusFound)
		})
	}
}
