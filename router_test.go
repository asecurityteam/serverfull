package serverfull

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRouterHasHealthCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fetcher := NewMockFetcher(ctrl)
	conf := &RouterConfig{
		Fetcher: fetcher,
	}
	router := NewRouter(conf)

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/healthcheck", http.NoBody)
	router.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)
}

func TestRouterHasLambdaInvoke(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fn := NewMockFunction(ctrl)
	fetcher := NewMockFetcher(ctrl)
	conf := &RouterConfig{
		Fetcher: fetcher,
	}
	router := NewRouter(conf)
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "http://localhost/2015-03-31/functions/TESTFUNCTION/invocations", http.NoBody)

	fetcher.EXPECT().Fetch(gomock.Any(), "TESTFUNCTION").Return(fn, nil)
	fn.EXPECT().Invoke(gomock.Any(), gomock.Any()).Return([]byte{}, nil)
	router.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)
}
