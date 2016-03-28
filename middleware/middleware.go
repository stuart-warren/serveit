package middleware

type Middleware func(http.Handler) http.Handler

func Decorate(h http.Handler, m ...Middleware) http.Handler {
	decorated := h
	// decorate is a function
	for _, decorate := range m {
		decorated = decorate(decorated)
	}
	return decorated
}

type Decorator func(Middleware) Middleware
