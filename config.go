package main

import (
	"fmt"
	"net"
)

type ServerType byte

const (
	NONE ServerType = iota
	STATIC
	PROXY
)

func ServerTypeParse(s string) (ServerType, error) {
	types := map[string]ServerType{"static": STATIC, "proxy": PROXY}
	if st, ok := types[s]; !ok {
		return NONE, fmt.Errorf("invalid Server type %q", s)
	} else {
		return st, nil
	}
}

type Config struct {
	Address     string
	IsTLS       bool
	TLSCertFile string
	TLSKeyFile  string
	ServerType  ServerType
	ProxyURL    string
	StaticDir   string
}

func Default(pref string, def string) string {
	if pref != "" {
		return pref
	}
	return def
}

func NewConfig(listen string) (Config, error) {
	serverType := STATIC
	if _, _, err := net.SplitHostPort(listen); err != nil {
		return Config{}, fmt.Errorf("invalid HOST:PORT parameter %s", listen)
	}
	return Config{Address: listen, ServerType: serverType}, nil
}
