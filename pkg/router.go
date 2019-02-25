package serverfull

import (
	"net/http"

	"github.com/asecurityteam/runhttp"
	"github.com/asecurityteam/serverfull/pkg/domain"
	v1 "github.com/asecurityteam/serverfull/pkg/handlers/v1"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// RouterConfig is used to alter the behavior of the default router
// and the HTTP endpoint handlers that it manages.
type RouterConfig struct {
	// HealthCheck defines the route on which the service will respond
	// with automatic 200s. This is here to integrate with systems that
	// poll for liveliness. The default value is /healthcheck
	HealthCheck string

	// HandlerFetcher is the Lambda function loader that will
	// be used by the runtime. There is no default for this value.
	HandlerFetcher domain.HandlerFetcher

	// LogFn is used to extract the request logger from the request
	// context. The default value is logevent.FromContext.
	LogFn domain.LogFn
	// StatFn is used to extract the request stat client from the
	// request context. The default value is xstats.FromContext.
	StatFn domain.StatFn
	// URLParamFn is used to extract URL parameters from the request.
	// The default value is chi.URLParam to match the usage of chi
	// as a mux in the default case.
	URLParamFn domain.URLParamFn
}

func applyDefaults(conf *RouterConfig) *RouterConfig {
	if conf.HealthCheck == "" {
		conf.HealthCheck = "/healthcheck"
	}
	if conf.LogFn == nil {
		conf.LogFn = runhttp.LoggerFromContext
	}
	if conf.StatFn == nil {
		conf.StatFn = runhttp.StatFromContext
	}
	if conf.URLParamFn == nil {
		conf.URLParamFn = chi.URLParamFromCtx
	}
	return conf
}

// NewRouter generates a mux that already has AWS Lambda API
// routes bound. This version returns a mux from the chi project
// as a convenience for cases where custom middleware or additional
// routes need to be configured.
func NewRouter(conf *RouterConfig) *chi.Mux {
	conf = applyDefaults(conf)
	router := chi.NewMux()
	router.Use(middleware.Heartbeat(conf.HealthCheck))

	invokeHandler := &v1.Invoke{
		Fetcher:    conf.HandlerFetcher,
		LogFn:      conf.LogFn,
		StatFn:     conf.StatFn,
		URLParamFn: conf.URLParamFn,
	}

	router.Method(http.MethodPost, "/2015-03-31/functions/{functionName}/invocations", invokeHandler)
	return router
}
