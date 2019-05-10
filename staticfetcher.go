package serverfull

import (
	"context"
)

// StaticFetcher is an implementation of the Fetcher that maintains a static mapping
// of names to Function instances. This implementation is a highly simplified form
// for the purposes of reducing risk in operations. Notably, runtimes that leverage
// this implementation do not need to perform any orchestration of external systems
// as all invocations of the Functions happen within the process and share the
// runtime's resources. Additionally, there is no "live update" feature which means
// there are less moving parts that might fail when attempting to start or update a
// Function.
//
// The trade-off is that updates to, additions of, and removals of Functions must be
// accomplished by generating a new build and redeploying the runtime. There are no
// options for updating or adding in-place. Operators and developers who choose this
// feature must take care that redeployments of the system do not cause downtime as
// all Functions will be affected together.
type StaticFetcher struct {
	// Functions is the underlying static map of function names to executable
	// functions. The keys of the map will be used as the name of the Function.
	Functions map[string]Function
}

// Fetch resolves the name using the internal mapping.
func (f *StaticFetcher) Fetch(ctx context.Context, name string) (Function, error) {
	h, ok := f.Functions[name]
	if !ok {
		return nil, NotFoundError{ID: name}
	}
	return h, nil
}
