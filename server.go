package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/bradfitz/http2"
)

type server struct {
	config Config
	logger *logrus.Logger
}

func NewServer(c Config) server {
	return server{config: c}
}

func (s *server) Run() error {
	server := &http.Server{
		Addr: s.config.Address,
		Handler: Decorate(
			defaultHandler(s.config),
			Logging(s.logger),
		)}
	http2.ConfigureServer(server, nil)

	if s.config.IsTLS {
		return server.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
	}
	return server.ListenAndServe()
}

func (s *server) SetLogger(logger *logrus.Logger) {
	s.logger = logger
}

func defaultHandler(c Config) http.Handler {
	wd, _ := os.Getwd()
	static := http.FileServer(http.Dir(Default(c.StaticDir, wd)))

	url, _ := url.Parse(Default(c.ProxyURL, "http://localhost:8000"))
	proxy := httputil.NewSingleHostReverseProxy(url)

	var st http.Handler
	switch c.ServerType {
	case PROXY:
		st = proxy
	case STATIC:
		st = static
	}

	return st
}
