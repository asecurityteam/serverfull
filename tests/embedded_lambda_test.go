// +build integration

package tests

// The following code is all copied directly from the lambda go SDK. Unfortunately,
// there is no easy way to alter the lambda server behavior in order to _stop_ a
// lambda once it is running. Perhaps AWS stops the servers by terminating the
// VM/container it is running in rather than through signal handling. The process
// of spinning up the server requires manipulation of unexported attributes of
// the Function type so we must embed it here so we can modify it.

// Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

type Function struct {
	handler lambda.Handler
}

func (fn *Function) Ping(req *messages.PingRequest, response *messages.PingResponse) error {
	*response = messages.PingResponse{}
	return nil
}

func (fn *Function) Invoke(req *messages.InvokeRequest, response *messages.InvokeResponse) error {
	defer func() {
		if err := recover(); err != nil {
			panicInfo := getPanicInfo(err)
			response.Error = &messages.InvokeResponse_Error{
				Message:    panicInfo.Message,
				Type:       getErrorType(err),
				StackTrace: panicInfo.StackTrace,
				ShouldExit: true,
			}
		}
	}()

	deadline := time.Unix(req.Deadline.Seconds, req.Deadline.Nanos).UTC()
	invokeContext, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	lc := &lambdacontext.LambdaContext{
		AwsRequestID:       req.RequestId,
		InvokedFunctionArn: req.InvokedFunctionArn,
		Identity: lambdacontext.CognitoIdentity{
			CognitoIdentityID:     req.CognitoIdentityId,
			CognitoIdentityPoolID: req.CognitoIdentityPoolId,
		},
	}
	if len(req.ClientContext) > 0 {
		if err := json.Unmarshal(req.ClientContext, &lc.ClientContext); err != nil {
			response.Error = lambdaErrorResponse(err)
			return nil
		}
	}
	invokeContext = lambdacontext.NewContext(invokeContext, lc)

	invokeContext = context.WithValue(invokeContext, "x-amzn-trace-id", req.XAmznTraceId)

	payload, err := fn.handler.Invoke(invokeContext, req.Payload)
	if err != nil {
		response.Error = lambdaErrorResponse(err)
		return nil
	}
	response.Payload = payload
	return nil
}

func getErrorType(err interface{}) string {
	errorType := reflect.TypeOf(err)
	if errorType.Kind() == reflect.Ptr {
		return errorType.Elem().Name()
	}
	return errorType.Name()
}

func lambdaErrorResponse(invokeError error) *messages.InvokeResponse_Error {
	var errorName string
	if errorType := reflect.TypeOf(invokeError); errorType.Kind() == reflect.Ptr {
		errorName = errorType.Elem().Name()
	} else {
		errorName = errorType.Name()
	}
	return &messages.InvokeResponse_Error{
		Message: invokeError.Error(),
		Type:    errorName,
	}
}

type panicInfo struct {
	Message    string                                      // Value passed to panic call, converted to string
	StackTrace []*messages.InvokeResponse_Error_StackFrame // Stack trace of the panic
}

func getPanicInfo(value interface{}) panicInfo {
	message := getPanicMessage(value)
	stack := getPanicStack()

	return panicInfo{Message: message, StackTrace: stack}
}

func getPanicMessage(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

var defaultErrorFrameCount = 32

func getPanicStack() []*messages.InvokeResponse_Error_StackFrame {
	s := make([]uintptr, defaultErrorFrameCount)
	const framesToHide = 3 // this (getPanicStack) -> getPanicInfo -> handler defer func
	n := runtime.Callers(framesToHide, s)
	if n == 0 {
		return make([]*messages.InvokeResponse_Error_StackFrame, 0)
	}

	s = s[:n]

	return convertStack(s)
}

func convertStack(s []uintptr) []*messages.InvokeResponse_Error_StackFrame {
	var converted []*messages.InvokeResponse_Error_StackFrame
	frames := runtime.CallersFrames(s)

	for {
		frame, more := frames.Next()

		formattedFrame := formatFrame(frame)
		converted = append(converted, formattedFrame)

		if !more {
			break
		}
	}
	return converted
}

func formatFrame(inputFrame runtime.Frame) *messages.InvokeResponse_Error_StackFrame {
	path := inputFrame.File
	line := int32(inputFrame.Line)
	label := inputFrame.Function

	// Strip GOPATH from path by counting the number of seperators in label & path
	//
	// For example given this:
	//     GOPATH = /home/user
	//     path   = /home/user/src/pkg/sub/file.go
	//     label  = pkg/sub.Type.Method
	//
	// We want to set:
	//     path  = pkg/sub/file.go
	//     label = Type.Method

	i := len(path)
	for n, g := 0, strings.Count(label, "/")+2; n < g; n++ {
		i = strings.LastIndex(path[:i], "/")
		if i == -1 {
			// Something went wrong and path has less seperators than we expected
			// Abort and leave i as -1 to counteract the +1 below
			break
		}
	}

	path = path[i+1:] // Trim the initial /

	// Strip the path from the function name as it's already in the path
	label = label[strings.LastIndex(label, "/")+1:]
	// Likewise strip the package name
	label = label[strings.Index(label, ".")+1:]

	return &messages.InvokeResponse_Error_StackFrame{
		Path:  path,
		Line:  line,
		Label: label,
	}
}

// StartHandler is slightly modified from the AWS lambda SDK. The server that the
// official SDK starts has no mechanism for stopping so we have to change it slightly
// for the test. Specifically, we need to add a signal handler and close the server.
// This opens us to the chance that the official SDK may skew from our tests.
func StartHandler(handler lambda.Handler) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	port := os.Getenv("_LAMBDA_SERVER_PORT")
	lis, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		log.Fatal(err)
	}
	function := &Function{handler}
	rpcServer := rpc.NewServer()
	err = rpcServer.Register(function)
	if err != nil {
		log.Fatal("failed to register handler function")
	}
	go rpcServer.Accept(lis)
	<-c
	_ = lis.Close()
}
