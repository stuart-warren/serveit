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
					jwt, err := base64.URLEncoding.DecodeString(cookie.Value)
					if err != nil {
						http.Error(w, "could not decode cookie", http.StatusBadRequest)
						return
					}
					tkn, err := oidcAuth.Verify(r.Context(), string(jwt))
					if err != nil {
						log.Printf("invalid jwt: %s", err)
						break
					}

					var claims struct {
						Subject       string `json:"sub"`
						Email         string `json:"email"`
						EmailVerified bool   `json:"email_verified"`
					}
					err = tkn.Claims(&claims)
					r.Header.Set("User", claims.Email)
					next.ServeHTTP(w, r)
					return

				}
			}
			for _, urlSuffix := range dontRedirect {
				if strings.HasSuffix(r.URL.Path, urlSuffix) {
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
