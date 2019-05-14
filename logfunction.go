package serverfull

import (
	"context"

	"github.com/asecurityteam/logevent"
)

type loggingFunction struct {
	Function
	Logger Logger
}

func (f *loggingFunction) Invoke(ctx context.Context, b []byte) ([]byte, error) {
	ctx = logevent.NewContext(ctx, f.Logger.Copy())
	return f.Function.Invoke(ctx, b)
}

// loggingFetcher wraps the function in a decorator that injects a logger.
type loggingFetcher struct {
	Logger  Logger
	Fetcher Fetcher
}

// Fetch calls the underlying Fetcher and adds log injection.
func (f *loggingFetcher) Fetch(ctx context.Context, name string) (Function, error) {
	r, err := f.Fetcher.Fetch(ctx, name)
	if err != nil {
		return nil, err
	}
	return &loggingFunction{Logger: f.Logger, Function: r}, nil
}
