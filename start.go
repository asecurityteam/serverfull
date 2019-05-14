package serverfull

import (
	"context"
	"fmt"
	"strings"

	"github.com/asecurityteam/runhttp"
	"github.com/asecurityteam/settings"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	// BuildModeHTTP is the standard mode of running an HTTP server
	// that implements parts of the Lambda API.
	BuildModeHTTP = "http"
	// BuildModeLambda runs the official lambda server using the lambda
	// SDK. Using this mode requires the TargetFunction value to be set.
	BuildModeLambda = "lambda"
)

var (
	// BuildMode determines the behavior of the Start method. There
	// are several ways to use this value. The suggested way is through
	// build variables by adding `-ldflags "-X github.com/asecurityteam/serverfull.BuildMode=<value>"`
	// to `go build` or `go run` commands. If you want to use environment variables
	// instead then you can set this variable in code before calling Start
	// like `serverfull.BuildMode=os.Getenv("MYENVVAR")`.
	//
	// Alternatively, the StartMode() or StartModeMock() method may be used if you
	// prefer to pass in parameters via code rather than toggling the global setting.
	BuildMode = BuildModeHTTP
	// MockMode determines whether or not to mock out the defined functions before
	// starting the server. Any non-empty string value will trigger mocking.
	MockMode = ""
	// TargetFunction is used when building in a native lambda mode to select a
	// single function to run. This value can be set in all the same ways as the
	// BuildMode value.
	TargetFunction = ""
)

// Start is a replacement for the lambda.Start method that introduces new
// features. By default, this method will start the lambda HTTP API and
// will invoke methods loaded using the given Fetcher.
func Start(ctx context.Context, s settings.Source, f Fetcher) error {
	if MockMode != "" {
		return StartModeMock(ctx, s, f, BuildMode, TargetFunction)
	}
	return StartMode(ctx, s, f, BuildMode, TargetFunction)
}

// StartMode works just like Start but allows for explicit passing of the build
// mode and target function.
func StartMode(ctx context.Context, s settings.Source, f Fetcher, mode string, target string) error {
	switch {
	case strings.EqualFold(mode, BuildModeHTTP):
		return StartHTTP(ctx, s, f)
	case strings.EqualFold(mode, BuildModeLambda):
		return StartLambda(ctx, s, f, target)
	default:
		return fmt.Errorf("unknown build mode %s", mode)
	}
}

// StartModeMock works just like StartMode but runs with mocked out
// functions.
func StartModeMock(ctx context.Context, s settings.Source, f Fetcher, mode string, target string) error {
	switch {
	case strings.EqualFold(mode, BuildModeHTTP):
		return StartHTTPMock(ctx, s, f)
	case strings.EqualFold(mode, BuildModeLambda):
		return StartLambdaMock(ctx, s, f, target)
	default:
		return fmt.Errorf("unknown build mode %s", mode)
	}
}

func newRuntime(ctx context.Context, s settings.Source, f Fetcher) (*runhttp.Runtime, error) {
	conf := &RouterConfig{
		Fetcher: f,
	}
	router := NewRouter(conf)
	rtC := &runhttp.Component{Handler: router}
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
	return StartHTTP(ctx, s, f)
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
