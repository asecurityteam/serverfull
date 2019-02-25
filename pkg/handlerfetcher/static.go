package handlerfetcher

import (
	"context"

	"github.com/asecurityteam/serverfull/pkg/domain"
)

// Static is an implementation of the HandlerFetcher that maintains a static mapping
// of names to Handler instances. This implementation is a highly simplified form
// for the purposes of reducing risk in operations. Notably, runtimes that leverage
// this implementation do not need to perform any orchestration of external systems
// as all invocations of the Handlers happen within the process and share the
// runtime's resources. Additionally, there is no "live update" feature which means
// there are less moving parts that might fail when attempting to start or update a
// Handler.
//
// The trade-off is that updates to, additions of, and removals of Handlers must be
// accomplished by generating a new build and redeploying the runtime. There are no
// options for updating or adding in-place. Operators and developers who choose this
// feature must take care that redeployments of the system do not cause downtime as
// all Handlers will be affected together.
type Static struct {
	// Handlers is the underlying static map of function names to executable
	// functions. The keys of the map will be used as the name of the Handler.
	Handlers map[string]domain.Handler
}

// FetchHandler resolves the name using the internal mapping.
func (f *Static) FetchHandler(ctx context.Context, name string) (domain.Handler, error) {
	h, ok := f.Handlers[name]
	if !ok {
		return nil, domain.NotFoundError{ID: name}
	}
	return h, nil
}
