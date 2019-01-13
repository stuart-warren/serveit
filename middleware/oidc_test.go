package middleware_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stuart-warren/serveit/middleware"
	"github.com/stuart-warren/serveit/oidc"
)

type MockVerifier struct{}

func (m MockVerifier) Verify(ctx context.Context, redirectTo string) (*oidc.IDToken, error) {
	return &oidc.IDToken{}, nil
}
func TestOIDCMiddleware(t *testing.T) {

	oidcAuth := MockVerifier{}
	oidcMiddleware := middleware.OIDC(middleware.NewOIDCMiddlewareConfig(oidcAuth))
	ts := httptest.NewServer(oidcMiddleware(GetTestHandler()))
	defer ts.Close()

	var u bytes.Buffer
	u.WriteString(string(ts.URL))
	u.WriteString("/")

	resp, err := http.Get(u.String())
	if err != nil {
		t.Error(err)
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%s", b)

	//FIXME this doesn't actually test anything
}
