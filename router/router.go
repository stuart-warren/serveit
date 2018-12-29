package router

import (
	"net/http"
	"regexp"
	"sync"

	"github.com/stuart-warren/serveit/access"
)

type Route interface {
	Match(path string) bool
	Permit(permitted access.Permitted) Route
	Permitted() access.Permitted
}

type textRoute struct {
	path      string
	permitted access.Permitted
}

func NewTextRoute(path string) Route {
	return textRoute{
		path:      path,
		permitted: access.Permitted{},
	}
}

func (t textRoute) Permit(permitted access.Permitted) Route {
	t.permitted = permitted
	return t
}

func (t textRoute) Permitted() access.Permitted {
	return t.permitted
}

func (t textRoute) Match(path string) bool {
	return path == t.path
}

func (t textRoute) String() string {
	return t.path
}

type regexRoute struct {
	pattern   *regexp.Regexp
	permitted access.Permitted
}

func NewRegexRoute(pattern *regexp.Regexp) Route {
	return regexRoute{
		pattern:   pattern,
		permitted: access.Permitted{},
	}
}

func (x regexRoute) Permit(permitted access.Permitted) Route {
	x.permitted = permitted
	return x
}

func (x regexRoute) Permitted() access.Permitted {
	return x.permitted
}

func (x regexRoute) Match(path string) bool {
	return x.pattern.MatchString(path)
}

func (x regexRoute) String() string {
	return x.pattern.String()
}

type prefixRoute struct {
	prefix    string
	permitted access.Permitted
}

func NewPrefixRoute(prefix string) Route {
	return prefixRoute{
		prefix:    prefix,
		permitted: access.Permitted{},
	}
}

func (p prefixRoute) Permit(permitted access.Permitted) Route {
	p.permitted = permitted
	return p
}

func (p prefixRoute) Permitted() access.Permitted {
	return p.permitted
}

func (p prefixRoute) Match(path string) bool {
	// strings.HasPrefix(s, prefix string) bool
	return len(path) >= len(p.prefix) && path[0:len(p.prefix)] == p.prefix
}

func (p prefixRoute) String() string {
	return p.prefix
}

type Router struct {
	mu         sync.RWMutex
	routes     []Route
	handler    http.Handler
	authorized func(http.ResponseWriter, *http.Request, Route) bool
}

func NewRouter(handler http.Handler, authorized func(w http.ResponseWriter, r *http.Request, route Route) bool) *Router {
	return &Router{
		routes:     []Route{},
		handler:    handler,
		authorized: authorized,
	}
}

func (o *Router) Reset() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.routes = []Route{}
}

func (o *Router) Handle(route Route) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.routes = append(o.routes, route)
}

func (o *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range o.routes {
		if route.Match(r.URL.Path) {
			if o.authorized(w, r, route) {
				o.handler.ServeHTTP(w, r)
				return
			} else {
				http.Error(w, "403 Forbidden", http.StatusForbidden)
				return
			}
		}
	}
	http.Error(w, "404 page not found", http.StatusNotFound)
}
