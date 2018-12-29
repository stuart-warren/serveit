package rules

import (
	"net/http"

	"github.com/stuart-warren/serveit/router"
)

var AllowAll = func(w http.ResponseWriter, r *http.Request, route router.Route) bool {
	return true
}

var DenyAll = func(w http.ResponseWriter, r *http.Request, route router.Route) bool {
	return false
}

var CheckMethod = func(w http.ResponseWriter, r *http.Request, route router.Route) bool {
	for _, m := range route.Permitted().Methods() {
		if m == "ALL" || m == r.Method {
			return true
		}
	}
	return false
}

var CheckUser = func(w http.ResponseWriter, r *http.Request, route router.Route) bool {
	users := route.Permitted().Users()
	for _, u := range users {
		// FIXME - insecure
		if u == "ALL" || u == r.Header.Get("User") {
			return true
		}
	}
	return false
}
