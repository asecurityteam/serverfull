package serverfull

import (
	"context"

	"github.com/asecurityteam/runhttp"
	"github.com/asecurityteam/serverfull/pkg/domain"
	"github.com/asecurityteam/serverfull/pkg/handlerfetcher"
	"github.com/asecurityteam/settings"
)

// NewStatic generates a runtime bound to the given handler mapping.
func NewStatic(ctx context.Context, s settings.Source, handlers map[string]domain.Handler) (*runhttp.Runtime, error) {
	fetcher := &handlerfetcher.Static{
		Handlers: handlers,
	}
	conf := &RouterConfig{
		HandlerFetcher: fetcher,
	}
	router := NewRouter(conf)
	rtC := &runhttp.Component{Handler: router}
	rt := new(runhttp.Runtime)
	err := settings.NewComponent(
		ctx,
		&settings.PrefixSource{Source: s, Prefix: []string{"SERVERFULL"}},
		rtC,
		rt,
	)
	return rt, err
}
