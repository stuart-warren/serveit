package oidc

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	oidc "github.com/coreos/go-oidc"
	"github.com/stuart-warren/serveit/encryption"
	"golang.org/x/oauth2"
)

const (
	NonceCookie    = "NONCE"
	JWTCookie      = "JWT"
	RedirectCookie = "REDIRECT"
	JWTHeader      = "X-JWT"
	StateStringFmt = "%v|%v"
)

type OIDCAuth struct {
	provider     *oidc.Provider
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
	cipher       encryption.Cipher
	secureCookie bool
	now          func() time.Time
}

type oidcAuthBuilder struct {
	cipher                                        encryption.Cipher
	clientID, clientSecret, redirectURL, provider string
	scopes                                        []string
	privateKey                                    []byte
	ctx                                           context.Context
	secureCookie                                  bool
	now                                           func() time.Time
}

type IDToken = oidc.IDToken

func NewOIDCAuth(clientID, clientSecret, redirectURL string) oidcAuthBuilder {
	return oidcAuthBuilder{
		ctx:          context.Background(),
		privateKey:   encryption.GenerateKey(),
		provider:     "https://accounts.google.com",
		scopes:       []string{"profile", "email"},
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		secureCookie: true,
		now:          time.Now,
	}
}

func (ob oidcAuthBuilder) WithContext(ctx context.Context) oidcAuthBuilder {
	ob.ctx = ctx
	return ob
}

func (ob oidcAuthBuilder) WithProvider(provider string) oidcAuthBuilder {
	ob.provider = provider
	return ob
}

func (ob oidcAuthBuilder) WithScopes(scopes []string) oidcAuthBuilder {
	ob.scopes = append(scopes, oidc.ScopeOpenID)
	return ob
}

func (ob oidcAuthBuilder) WithInsecureCookies() oidcAuthBuilder {
	ob.secureCookie = false
	return ob
}

func (ob oidcAuthBuilder) WithNowFunc(now func() time.Time) oidcAuthBuilder {
	ob.now = now
	return ob
}

// WithPrivate key takes a 32bit key from encryption.GenerateKey()
func (ob oidcAuthBuilder) WithPrivateKey(key []byte) oidcAuthBuilder {
	ob.privateKey = key
	return ob
}

func (ob oidcAuthBuilder) Build() (OIDCAuth, error) {
	provider, err := oidc.NewProvider(ob.ctx, ob.provider)
	if err != nil {
		return OIDCAuth{}, err
	}
	return OIDCAuth{
		provider: provider,
		oauth2Config: oauth2.Config{
			ClientID:     ob.clientID,
			ClientSecret: ob.clientSecret,
			RedirectURL:  ob.redirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       ob.scopes,
		},
		verifier: provider.Verifier(&oidc.Config{ClientID: ob.clientID}),
		cipher:   encryption.NewMiscreantCipher(ob.privateKey),
	}, nil
}

// HandleRedirect creates a session nonce, encrypts it in a cookie, builds a secret state and redirects user to auth service
func (o OIDCAuth) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	nonce := fmt.Sprintf("%x", encryption.GenerateKey())
	http.SetCookie(w, &http.Cookie{
		Name:     NonceCookie,
		Value:    base64.URLEncoding.EncodeToString(must(o.cipher.Encrypt([]byte(nonce)))),
		Expires:  o.now().Add(5 * time.Minute),
		Secure:   o.secureCookie,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		HttpOnly: true,
	})
	state := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf(StateStringFmt, nonce, o.oauth2Config.RedirectURL)))
	http.Redirect(w, r, o.oauth2Config.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound)
}

// HandleCallBack uses cookie to rebuild secret state variable to validate user has been authed
func (o OIDCAuth) HandleCallBack(w http.ResponseWriter, r *http.Request) {
	var err error
	var nonce []byte
	redirectTo := "/"
	for _, cookie := range r.Cookies() {
		if cookie.Name == NonceCookie {
			val, err := base64.URLEncoding.DecodeString(cookie.Value)
			if err != nil {
				http.Error(w, "could not decode cookie", http.StatusBadRequest)
				return
			}
			nonce, err = o.cipher.Decrypt(val)
			if err != nil {
				http.Error(w, "could not decrypt cookie", http.StatusBadRequest)
				return
			}
		}
		if cookie.Name == RedirectCookie {
			val, err := base64.URLEncoding.DecodeString(cookie.Value)
			if err != nil {
				http.Error(w, "could not decode cookie", http.StatusBadRequest)
				return
			}
			redirectTo = string(val)
		}
	}
	if len(nonce) == 0 {
		http.Error(w, "missing nonce cookie", http.StatusBadRequest)
		return
	}
	expectedState := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf(StateStringFmt, string(nonce), o.oauth2Config.RedirectURL)))
	if r.URL.Query().Get("state") != expectedState {
		http.Error(w, "state did not match or is missing", http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	oauth2Token, err := o.oauth2Config.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// userInfo, err := o.provider.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token))
	// if err != nil {
	// 	http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}
	idToken, err := o.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if idToken.Nonce != string(nonce) {
		http.Error(w, "invalid ID Token nonce", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     JWTCookie,
		Value:    base64.URLEncoding.EncodeToString([]byte(rawIDToken)),
		Expires:  idToken.Expiry,
		Secure:   o.secureCookie,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		HttpOnly: true,
	})
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

func (o OIDCAuth) Verify(ctx context.Context, rawIDToken string) (*IDToken, error) {
	return o.verifier.Verify(ctx, rawIDToken)
}

func must(data []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return data
}
