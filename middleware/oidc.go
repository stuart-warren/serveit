package middleware

import (
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/stuart-warren/serveit/oidc"
)

func OIDC(oidcAuth oidc.OIDCAuth, redirectTo string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &http.Cookie{
				Name:     oidc.RedirectCookie,
				Value:    base64.URLEncoding.EncodeToString([]byte(r.RequestURI)),
				Expires:  time.Now().Add(1 * time.Minute),
				Secure:   false, //FIXME
				SameSite: http.SameSiteStrictMode,
				Path:     "/",
				HttpOnly: true,
			})
			shouldRedirectToAuth := true
			for _, cookie := range r.Cookies() {
				if cookie.Name == oidc.JWTCookie {
					jwt, err := base64.URLEncoding.DecodeString(cookie.Value)
					if err != nil {
						http.Error(w, "could not decode cookie", http.StatusBadRequest)
						return
					}
					_, err = oidcAuth.Verify(r.Context(), string(jwt))
					if err != nil {
						log.Printf("invalid jwt")
						http.Redirect(w, r, redirectTo, http.StatusFound)
						return
					}
					log.Printf("valid jwt")
					shouldRedirectToAuth = false
				}
				if cookie.Name == oidc.RedirectCookie {
					log.Printf("has redirect cookie")
					shouldRedirectToAuth = false
				}
			}
			if shouldRedirectToAuth {
				log.Printf("redirecting because i should")
				http.Redirect(w, r, redirectTo, http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
