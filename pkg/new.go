package serverfull

import (
	"context"
	"os"

	"github.com/asecurityteam/runhttp"
	"github.com/asecurityteam/serverfull/pkg/domain"
	"github.com/asecurityteam/serverfull/pkg/handlerfetcher"
	"github.com/asecurityteam/settings"
)

// NewStatic generates a runtime bound to the given handler mapping.
func NewStatic(handlers map[string]domain.Handler) (*runhttp.Runtime, error) {
	fetcher := &handlerfetcher.Static{
		Handlers: handlers,
	}
	conf := &RouterConfig{
		HandlerFetcher: fetcher,
	}
	router := NewRouter(conf)
	source, _ := settings.NewEnvSource(os.Environ())
	rtC := &runhttp.Component{Handler: router}
	rt := new(runhttp.Runtime)
	err := settings.NewComponent(
		context.Background(),
		&settings.PrefixSource{Source: source, Prefix: []string{"SERVERFULL"}},
		rtC,
		rt,
	)
	return rt, err
}
