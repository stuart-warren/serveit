package middleware

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/stuart-warren/serveit/oidc"
)

type Verifier interface {
	Verify(ctx context.Context, redirectTo string) (*oidc.IDToken, error)
}

func OIDC(oidcAuth Verifier, redirectTo string, dontRedirect []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, cookie := range r.Cookies() {
				if cookie.Name == oidc.JWTCookie {
					log.Printf("has jwt cookie")
					jwt, err := base64.URLEncoding.DecodeString(cookie.Value)
					if err != nil {
						http.Error(w, "could not decode cookie", http.StatusBadRequest)
						return
					}
					_, err = oidcAuth.Verify(r.Context(), string(jwt))
					if err != nil {
						log.Printf("invalid jwt: %s", err)
					}
					log.Printf("valid jwt")
					next.ServeHTTP(w, r)
					return
				}
			}
			// No jwt cookie found
			for _, urlSuffix := range dontRedirect {
				if strings.HasSuffix(r.URL.Path, urlSuffix) {
					log.Printf("no need to redirect")
					next.ServeHTTP(w, r)
					return
				}
			}
			http.SetCookie(w, &http.Cookie{
				Name:     oidc.RedirectCookie,
				Value:    base64.URLEncoding.EncodeToString([]byte(r.RequestURI)),
				Expires:  time.Now().Add(1 * time.Minute),
				Secure:   false, //FIXME
				SameSite: http.SameSiteStrictMode,
				Path:     "/",
				HttpOnly: true,
			})
			http.Redirect(w, r, redirectTo, http.StatusFound)
		})
	}
}
