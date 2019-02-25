package domain

import (
	"context"
)

// HandlerFetcher is a pluggable component that enables different
// loading strategies functions.
type HandlerFetcher interface {
	// FetchHandler uses some implementation of a loading strategy
	// to fetch the Handler with the given name. If a matching Handler
	// cannot be found then this component must emit a NotFoundError.
	FetchHandler(ctx context.Context, name string) (Handler, error)
}
