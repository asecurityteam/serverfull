// +build integration

package tests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"

	serverfull "github.com/asecurityteam/serverfull/pkg"
	"github.com/asecurityteam/serverfull/pkg/domain"
	"github.com/asecurityteam/serverfull/pkg/handlerfetcher"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stretchr/testify/assert"
)

func TestURLRouting(t *testing.T) {
	fetcher := &handlerfetcher.Static{
		Handlers: map[string]domain.Handler{
			"hello": lambda.NewHandler(func() (string, error) { return "Hello ƛ!", nil }),
		},
	}
	conf := &serverfull.RouterConfig{
		HandlerFetcher: fetcher,
	}
	router := serverfull.NewRouter(conf)
	server := httptest.NewServer(router)
	defer server.Close()

	u, _ := url.Parse(server.URL)
	u.Path = path.Join(u.Path, "2015-03-31", "functions", "hello", "invocations")
	req, _ := http.NewRequest(http.MethodPost, u.String(), http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	u, _ = url.Parse(server.URL)
	u.Path = path.Join(u.Path, "healthcheck")
	req, _ = http.NewRequest(http.MethodGet, u.String(), http.NoBody)
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
