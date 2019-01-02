package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/stuart-warren/serveit/middleware"
	"github.com/stuart-warren/serveit/oidc"
)

var (
	// Create Google web application clientIDs from https://console.developers.google.com/apis/credentials
	clientID     = os.Getenv("OAUTH2_PROXY_CLIENT_ID")
	clientSecret = os.Getenv("OAUTH2_PROXY_CLIENT_SECRET")
	redirectURL  = os.Getenv("OAUTH2_REDIRECT_URL")
)

func main() {
	mux := http.NewServeMux()
	oidcAuth, err := oidc.NewOIDCAuth(clientID, clientSecret, redirectURL).Build()
	if err != nil {
		log.Fatal(err)
	}
	proxyURL, err := url.ParseRequestURI("http://localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	mux.HandleFunc("/callback", oidcAuth.HandleCallBack)
	mux.HandleFunc("/auth", oidcAuth.HandleRedirect)
	mux.HandleFunc("/", httputil.NewSingleHostReverseProxy(proxyURL).ServeHTTP)
	srv := &http.Server{Addr: ":1234", Handler: middleware.Decorate(mux, middleware.Logging(), middleware.OIDC(oidcAuth, "/auth"))}
	log.Printf("starting at %s\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
