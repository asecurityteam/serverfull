// +build integration

package tests

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/asecurityteam/serverfull"
	"github.com/asecurityteam/settings"
	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/stretchr/testify/require"
)

// hello is lifted straight from the aws-lambda-go README.md file.
// This is can be called like:
//
//		curl --request POST localhost:8080/2015-03-31/functions/hello/invocations
func hello() (string, error) {
	return "Hello ƛ!", nil
}

type logLine struct {
	Message string `logevent:"message,default=log-line"`
}

// logstat guarantees that the log and stat functions do not panic at runtime.
func logstat(ctx context.Context) {
	serverfull.LoggerFromContext(ctx).Info(logLine{})
	serverfull.StatFromContext(ctx).Count("stat", 1)
}

func TestStart(t *testing.T) {
	ctx := context.Background()
	functions := map[string]serverfull.Function{
		"hello":   serverfull.NewFunction(hello),
		"logstat": serverfull.NewFunction(logstat),
		"error":   serverfull.NewFunctionWithErrors(hello, errors.New("mock mode")),
	}
	fetcher := &serverfull.StaticFetcher{Functions: functions}
	// These tests are not safe to run in parallel but the subtest is parallel
	// by default unless we modify the `go test` command to include special values.
	// To work around this we've introduces a mutex to ensure only one test is running
	// concurrently. Ordering of the tests does not matter.
	mut := &sync.Mutex{}

	// makeHTTPCall attempts to execute the lambda over the invoke API until
	// either a success case is found or the loop times out.
	var makeHTTPCall = func(t *testing.T, port string) error {
		// Ping the server until it is available or until we exceed a timeout
		// value. This is to account for arbitrary start-up time of the server
		// in the background.
		stop := time.Now().Add(5 * time.Second)
		for time.Now().Before(stop) {
			time.Sleep(100 * time.Millisecond)
			resp, err := http.DefaultClient.Post(
				fmt.Sprintf(
					"http://localhost:%s/2015-03-31/functions/error/invocations",
					port,
				),
				"application/json",
				http.NoBody,
			)
			if err != nil {
				t.Log(err.Error())
				continue
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				b, _ := ioutil.ReadAll(resp.Body)
				t.Log(resp.StatusCode)
				t.Log(string(b))
				continue
			}
			return nil
		}
		return errors.New("failed to execute function")
	}

	// makeHTTPErrorCall attempts to execute the mock error simulation.
	var makeHTTPErrorCall = func(t *testing.T, port string) error {
		// Ping the server until it is available or until we exceed a timeout
		// value. This is to account for arbitrary start-up time of the server
		// in the background.
		stop := time.Now().Add(5 * time.Second)
		for time.Now().Before(stop) {
			time.Sleep(100 * time.Millisecond)
			req, _ := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf(
					"http://localhost:%s/2015-03-31/functions/error/invocations",
					port,
				),
				http.NoBody,
			)
			req.Header.Set("X-Amz-Invocation-Type", "Error")
			req.Header.Set("X-Error-Type", "errorString")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Log(err.Error())
				continue
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusInternalServerError {
				b, _ := ioutil.ReadAll(resp.Body)
				t.Log(resp.StatusCode)
				t.Log(string(b))
				continue
			}
			return nil
		}
		return errors.New("failed to execute function")
	}

	// Patch the lambda server so we can handle signals.
	originalStartHandler := serverfull.LambdaStartFn
	defer func() {
		serverfull.LambdaStartFn = originalStartHandler
	}()
	serverfull.LambdaStartFn = StartHandler

	// makeRPCCall imitates the internal execution path for the native lambda
	// system by using the net/rpc module.
	var makeRPCCall = func(t *testing.T, port string) error {
		// Ping the server until it is available or until we exceed a timeout
		// value. This is to account for arbitrary start-up time of the server
		// in the background.
		stop := time.Now().Add(5 * time.Second)
		for time.Now().Before(stop) {
			time.Sleep(100 * time.Millisecond)
			client, err := rpc.Dial("tcp", "localhost:"+port)
			if err != nil {
				t.Log(err.Error())
				continue
			}
			req := &messages.InvokeRequest{
				Payload: []byte(`{}`),
			}
			res := &messages.InvokeResponse{}
			if err := client.Call("Function.Invoke", req, res); err != nil {
				t.Log(err.Error())
				continue
			}
			if res.Error != nil {
				t.Log(res.Error.Message)
				continue
			}
			return nil
		}
		return errors.New("failed to execute function")
	}

	for _, testCase := range []struct {
		BuildMode      string
		MockMode       string
		TargetFunction string
		Execute        func(t *testing.T, port string) error
	}{
		{
			BuildMode:      serverfull.BuildModeHTTP,
			MockMode:       "true",
			TargetFunction: "error",
			Execute:        makeHTTPErrorCall,
		},
		{
			BuildMode:      serverfull.BuildModeHTTP,
			MockMode:       "",
			TargetFunction: "hello",
			Execute:        makeHTTPCall,
		},
		{
			BuildMode:      serverfull.BuildModeHTTP,
			MockMode:       "true",
			TargetFunction: "hello",
			Execute:        makeHTTPCall,
		},
		{
			BuildMode:      serverfull.BuildModeLambda,
			MockMode:       "",
			TargetFunction: "hello",
			Execute:        makeRPCCall,
		},
		{
			BuildMode:      serverfull.BuildModeLambda,
			MockMode:       "true",
			TargetFunction: "hello",
			Execute:        makeRPCCall,
		},
		{
			BuildMode:      serverfull.BuildModeHTTP,
			MockMode:       "",
			TargetFunction: "logstat",
			Execute:        makeHTTPCall,
		},
		{
			BuildMode:      serverfull.BuildModeHTTP,
			MockMode:       "true",
			TargetFunction: "logstat",
			Execute:        makeHTTPCall,
		},
		{
			BuildMode:      serverfull.BuildModeLambda,
			MockMode:       "",
			TargetFunction: "logstat",
			Execute:        makeRPCCall,
		},
		{
			BuildMode:      serverfull.BuildModeLambda,
			MockMode:       "true",
			TargetFunction: "logstat",
			Execute:        makeRPCCall,
		},
	} {
		t.Run(testCase.BuildMode+"/"+testCase.MockMode+"/"+testCase.TargetFunction, func(t *testing.T) {
			mut.Lock()
			defer mut.Unlock()

			port, err := getPort()
			require.NoError(t, err)

			// The native lambda function defines and manages its own set of environment
			// variables that we can't patch or mock out other than setting them for the
			// duration of the test. This variable defines the listening port for the RPC
			// server.
			os.Setenv("_LAMBDA_SERVER_PORT", port)

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
				"SERVERFULL_RUNTIME_HTTPSERVER_ADDRESS=localhost:" + port,
				"SERVERFULL_RUNTIME_LOGGER_OUTPUT=NULL",
				"SERVERFULL_RUNTIME_STATS_OUTPUT=NULL",
			})
			require.Nil(t, err)

			serverfull.BuildMode = testCase.BuildMode
			serverfull.MockMode = testCase.MockMode
			serverfull.TargetFunction = testCase.TargetFunction
			exit := make(chan error)
			go func() {
				exit <- serverfull.Start(ctx, source, fetcher)
			}()
			require.NoError(t, testCase.Execute(t, port))
			// The runtime establishes a signal handler for the entire
			// process. This means we have the process signal itself and
			// the runtime will intercept the call. This enables us to test
			// the signal based shutdown behavior.
			proc, _ := os.FindProcess(os.Getpid())
			_ = proc.Signal(os.Interrupt)
			select {
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for exit")
			case err := <-exit:
				require.Nil(t, err)
			}
		})
	}
}
