// +build integration

package tests

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/asecurityteam/serverfull/pkg"
	"github.com/asecurityteam/serverfull/pkg/domain"
	"github.com/asecurityteam/settings"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stretchr/testify/require"
)

// hello is lifted straight from the aws-lambda-go README.md file.
// This is can be called like:
//
//		curl --request POST localhost:8080/2015-03-31/functions/hello/invocations
func hello() (string, error) {
	return "Hello ƛ!", nil
}

func TestNew(t *testing.T) {
	ctx := context.Background()
	handlers := map[string]domain.Handler{
		"hello": lambda.NewHandler(hello),
	}

	// Rather than mock out the settings.Source, it ends up being easier
	// to manage and slightly more realistic to use the ENV source but
	// populated with a static ENV list. This is easier because we don't
	// need to mock out the internal call structure of the settings.Source
	// which is largely irrelevant to this test. This is more realistic
	// because it leverages the public configuration API of the project
	// rather than internal knowledge of the settings project. For example,
	// these ENV vars are exactly the ones that users would set when running
	// the system.
	source, err := settings.NewEnvSource([]string{
		"SERVERFULL_RUNTIME_HTTPSERVER_ADDRESS=localhost:9090",
		"SERVERFULL_RUNTIME_LOGGER_OUTPUT=NULL",
		"SERVERFULL_RUNTIME_STATS_OUTPUT=NULL",
	})
	require.Nil(t, err)
	rt, err := serverfull.NewStatic(ctx, source, handlers)
	require.Nil(t, err)

	exit := make(chan error)
	go func() {
		exit <- rt.Run()
	}()

	// Ping the server until it is available or until we exceed a timeout
	// value. This is to account for arbitrary start-up time of the server
	// in the background.
	stop := time.Now().Add(5 * time.Second)
	for time.Now().Before(stop) {
		time.Sleep(100 * time.Millisecond)
		resp, err := http.DefaultClient.Post(
			"http://localhost:9090/2015-03-31/functions/hello/invocations",
			"application/json",
			http.NoBody,
		)
		if err != nil {
			t.Log(err.Error())
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Log(resp.StatusCode)
			continue
		}
		break
	}
	// The runtime establishes a signal handler for the entire
	// process. This means we have the process signal itself and
	// the runtime will intercept the call. This enables us to test
	// the signal based shutdown behavior.
	proc, _ := os.FindProcess(os.Getpid())
	proc.Signal(os.Interrupt)
	select {
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for exit")
	case err := <-exit:
		require.Nil(t, err)
	}
}
