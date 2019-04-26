package serverfull

import (
	"context"
	"fmt"
	"strings"

	"github.com/asecurityteam/runhttp"
	"github.com/asecurityteam/settings"
)

const (
	// BuildModeHTTP is the standard mode of running an HTTP server
	// that implements parts of the Lambda API.
	BuildModeHTTP = "http"
	// BuildModeHTTPMock runs the HTTP server but with mocked versions
	// of the lambda functions loaded.
	BuildModeHTTPMock = "http_mock"
	// BuildModeLambda runs the official lambda server using the lambda
	// SDK. Using this mode requires the TargetFunction value to be set.
	BuildModeLambda = "lambda"
	// BuildModeLambdaMock runs the official lambda server using the lambda
	// SDK but with a mocked version of the loaded function. Using this mode
	// requires the TargetFunction value to be set.
	BuildModeLambdaMock = "lambda_mock"
)

var (
	// BuildMode determines the behavior of the Start method. There
	// are several ways to use this value. The suggested way is through
	// build variables by adding `-ldflags "-X github.com/asecurityteam.BuildMode=<value>"`
	// to `go build` or `go run` commands. If you want to use environment variables
	// instead then you can set this variable in code before calling Start
	// like `serverfull.BuildMode=os.Getenv("MYENVVAR")`.
	//
	// Alternatively, the StartMode() method may be used if you prefer to pass in
	// parameters via code rather than toggling the global setting.
	BuildMode = BuildModeHTTP
	// TargetFunction is used when building in a native lambda mode to select a
	// single function to run. This value can be set in all the same ways as the
	// BuildMode value.
	TargetFunction = ""
)

// Start is a replacement for the lambda.Start method that introduces new
// features. By default, this method will start the lambda HTTP API and
// will invoke methods loaded using the given Fetcher.
func Start(ctx context.Context, s settings.Source, f Fetcher) error {
	return StartMode(ctx, s, f, BuildMode, TargetFunction)
}

// StartMode works just like Start but allows for explicit passing of the build
// mode and target function.
func StartMode(ctx context.Context, s settings.Source, f Fetcher, mode string, target string) error {
	switch {
	case strings.EqualFold(mode, BuildModeHTTP):
		return StartHTTP(ctx, s, f)
	case strings.EqualFold(mode, BuildModeHTTPMock):
		return StartHTTPMock(ctx, s, f)
	case strings.EqualFold(mode, BuildModeLambda):
		// return StartLambda(ctx, s, f, target)
	case strings.EqualFold(mode, BuildModeLambdaMock):
		// return StartLambdaMock(ctx, s, f, target)
	default:
		return fmt.Errorf("unknown build mode %s", mode)
	}
	return fmt.Errorf("TODO: Complete the rest of the cases.")
}

func newHTTPRuntime(ctx context.Context, s settings.Source, f Fetcher) (*runhttp.Runtime, error) {
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
	rt, err := newHTTPRuntime(ctx, s, f)
	if err != nil {
		return err
	}
	return rt.Run()
}

// StartHTTP runs the HTTP API with mocked out functions.
func StartHTTPMock(ctx context.Context, s settings.Source, f Fetcher) error {
	f = &MockingFetcher{Fetcher: f}
	return StartHTTPMock(ctx, s, f)
}
