//go:build integration
// +build integration

package tests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/asecurityteam/serverfull"
)

func TestURLRouting(t *testing.T) {
	fetcher := &serverfull.StaticFetcher{
		Functions: map[string]serverfull.Function{
			"hello": serverfull.NewFunction(func() (string, error) { return "Hello ƛ!", nil }),
		},
	}
	conf := &serverfull.RouterConfig{
		Fetcher: fetcher,
	}
	router := serverfull.NewRouter(conf)
	server := httptest.NewServer(router)
	defer server.Close()

	u, _ := url.Parse(server.URL)
	u.Path = path.Join(u.Path, "2015-03-31", "functions", "hello", "invocations")
	req, _ := http.NewRequest(http.MethodPost, u.String(), http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	u, _ = url.Parse(server.URL)
	u.Path = path.Join(u.Path, "healthcheck")
	req, _ = http.NewRequest(http.MethodGet, u.String(), http.NoBody)
	resp, err = http.DefaultClient.Do(req)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
