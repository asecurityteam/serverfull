package domain

import (
	"context"
)

// URLParamFn should be accepted by HTTP handlers that need
// to interface with the mux in use in order to extract request
// parameters from the URL. This defines the contract between
// any given mux and a handler so that the two do not need to
// be coupled.
type URLParamFn func(ctx context.Context, name string) string
