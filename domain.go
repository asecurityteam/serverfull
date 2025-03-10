package serverfull

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/xstats"

	"github.com/asecurityteam/logevent/v2"
)

// Logger is an alias for the chosen project logging library
// which is, currently, logevent. All references in the project
// should be to this name rather than logevent directly.
type Logger = logevent.Logger

// LogFn extracts a logger from the context.
type LogFn = func(context.Context) Logger

// LoggerFromContext extracts the current logger.
var LoggerFromContext = logevent.FromContext

// Stat is an alias for the chosen project metrics library
// which is, currently, xstats. All references in the project
// should be to this name rather than xstats directly.
type Stat = xstats.XStater

// StatFn extracts a metrics client from the context.
type StatFn = func(context.Context) Stat

// StatFromContext extracts the current stat client.
var StatFromContext = xstats.FromContext

// Function is an executable lambda function. This extends
// the official lambda SDK concept of a Handler in order to
// also provide the underlying function signature which is
// usually masked when converting any function to a lambda.Handler.
type Function interface {
	lambda.Handler
	Source() interface{}
	Errors() []error
}

// URLParamFn should be accepted by HTTP handlers that need
// to interface with the mux in use in order to extract request
// parameters from the URL. This defines the contract between
// any given mux and a handler so that the two do not need to
// be coupled.
type URLParamFn func(ctx context.Context, name string) string

// Fetcher is a pluggable component that enables different
// loading strategies functions.
type Fetcher interface {
	// Fetch uses some implementation of a loading strategy
	// to fetch the Handler with the given name. If a matching Handler
	// cannot be found then this component must emit a NotFoundError.
	Fetch(ctx context.Context, name string) (Function, error)
}

// NotFoundError represents a failed lookup for a resource.
type NotFoundError struct {
	// ID is the key used when looking for the resource.
	ID string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("resource (%s) not found", e.ID)
}
