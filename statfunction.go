package serverfull

import (
	"context"

	"github.com/rs/xstats"
)

type statFunction struct {
	Function
	Stat Stat
}

func (f *statFunction) Invoke(ctx context.Context, b []byte) ([]byte, error) {
	ctx = xstats.NewContext(ctx, f.Stat)
	return f.Function.Invoke(ctx, b)
}

// statFetcher wraps the function in a decorator that injects a stat client.
type statFetcher struct {
	Stat    Stat
	Fetcher Fetcher
}

// Fetch calls the underlying Fetcher and adds stat client injection.
func (f *statFetcher) Fetch(ctx context.Context, name string) (Function, error) {
	r, err := f.Fetcher.Fetch(ctx, name)
	if err != nil {
		return nil, err
	}
	return &statFunction{Stat: f.Stat, Function: r}, nil
}
