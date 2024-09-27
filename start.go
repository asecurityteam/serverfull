package serverfull

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/asecurityteam/runhttp"
	"github.com/asecurityteam/settings/v2"
)

// Start is maintained for backwards compatibility but is deprecated.
// The original intent was a single entry point that would switch behaviors
// based on a build flag. However, build flags are not persisted after the build
// so the build-time selection is not actually available for switching at
// runtime. As a result, this function now only starts the HTTP server.
func Start(ctx context.Context, s settings.Source, f Fetcher) error {
	return StartHTTP(ctx, s, f)
}

func newRuntime(ctx context.Context, s settings.Source, f Fetcher) (*runhttp.Runtime, error) {
	conf := &RouterConfig{
		Fetcher: f,
	}
	router := NewRouter(conf)
	rtC := runhttp.NewComponent().WithHandler(router)
	rt := new(runhttp.Runtime)
	err := settings.NewComponent(
		ctx,
		&settings.PrefixSource{Source: s, Prefix: []string{"serverfull"}},
		rtC,
		rt,
	)
	return rt, err
}

func newMockRuntime(ctx context.Context, s settings.Source, f Fetcher) (*runhttp.Runtime, error) {
	conf := &RouterConfig{
		Fetcher:  f,
		MockMode: true,
	}
	router := NewRouter(conf)
	rtC := runhttp.NewComponent().WithHandler(router)
	rt := new(runhttp.Runtime)
	err := settings.NewComponent(
		ctx,
		&settings.PrefixSource{Source: s, Prefix: []string{"serverfull"}},
		rtC,
		rt,
	)
	return rt, err
}

// StartHTTP runs the HTTP API.
func StartHTTP(ctx context.Context, s settings.Source, f Fetcher) error {
	rt, err := newRuntime(ctx, s, f)
	if err != nil {
		return err
	}
	return rt.Run()
}

// StartHTTPMock runs the HTTP API with mocked out functions.
func StartHTTPMock(ctx context.Context, s settings.Source, f Fetcher) error {
	f = &MockingFetcher{Fetcher: f}
	rt, err := newMockRuntime(ctx, s, f)
	if err != nil {
		return err
	}
	return rt.Run()
}

// LambdaStartFn is a reference to lambda.StartHandler that is exported
// for cases where a custom net/rpc server needs to run rather than the
// true native lambda server. For example, this project leverages this
// feature in our integration tests where we add some additional signal
// handling for testing purposes.
var LambdaStartFn = lambda.StartHandler

// StartLambda runs the target function from the fetcher as a
// native lambda server.
func StartLambda(ctx context.Context, s settings.Source, f Fetcher, target string) error {
	rt, err := newRuntime(ctx, s, f)
	if err != nil {
		return err
	}
	// We hit an edge case in the go type system as it relates to type aliases.
	// The runhttp alias of `github.com/asecurityteam/logevent` resolves as
	// exactly that: `github.com/asecurityteam/logevent`. However, our own alias
	// here of `type Logger = logevent.Logger` actually resolves to
	// `github.com/asecurityteam/serverfull/vendor/github.com/asecurityteam/logevent`
	// which causes the compiler to error because the types are prefixed with
	// different package names. These two types are exactly the same but the
	// compiler is unable to figure this out. As a result we must erase the
	// compiler's knowledge of the type by switching to empty interface and then
	// re-type the value as our Logger.
	var typeHack interface{} = rt.Logger
	f = &loggingFetcher{Fetcher: f, Logger: typeHack.(Logger)}
	typeHack = rt.Stats
	f = &statFetcher{Fetcher: f, Stat: typeHack.(Stat)}
	fn, err := f.Fetch(ctx, target)
	if err != nil {
		return err
	}
	LambdaStartFn(fn)
	return nil
}

// StartLambdaMock starts the native lambda server with a mocked out
// function.
func StartLambdaMock(ctx context.Context, s settings.Source, f Fetcher, target string) error {
	f = &MockingFetcher{Fetcher: f}
	return StartLambda(ctx, s, f, target)
}
