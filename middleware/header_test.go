package middleware_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stuart-warren/serveit/middleware"
)

// https://medium.com/@PurdonKyle/unit-testing-golang-http-middleware-c7727ca896ea
func TestHeaderMiddleware(t *testing.T) {

	responseHeader := middleware.ResponseHeader("TEST", "FOOBAR")
	ts := httptest.NewServer(responseHeader(GetTestHandler()))
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

	for k, vl := range resp.Header {
		v := vl[0]
		if k == "TEST" {
			if v != "FOOBAR" {
				t.Errorf("got %s=%s instead", k, v)
			}
		}
		t.Logf("got %s=%s", k, v)
	}
}
