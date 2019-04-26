package serverfull

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
)

const (
	invocationTypeHeader          = "X-Amz-Invocation-Type"
	invocationTypeRequestResponse = "RequestResponse"
	invocationTypeEvent           = "Event"
	invocationTypeDryRun          = "DryRun"
	invocationVersionHeader       = "X-Amz-Executed-Version"
	invocationErrorHeader         = "X-Amz-Function-Error"
	invocationErrorTypeHandled    = "Handled"
	invocationErrorTypeUnhandled  = "Unhandled"
)

// bgContext is used to detach the *http.Request context from the http.Handler
// lifecycle. Typically, the request context is canceled when the hander returns.
// This is problematic when using the request context to share request scoped
// elements, such as the logger or stat client, with background tasks that will
// execute after the handler returns. This resolves that issue by keeping a
// reference to the request context and using it to lookup values but replacing
// all other context.Context methods with the context.Background() implementation.
// The result is a valid context.Context that will not expire when the source
// http.Handler returns but will maintain all context values.
type bgContext struct {
	context.Context
	Values context.Context
}

func (c *bgContext) Value(key interface{}) interface{} {
	return c.Values.Value(key)
}

// lambdaError implements the common Lambda error response
// JSON object that is included as the response body for
// exception cases.
type lambdaError struct {
	Message    string   `json:"errorMessage"`
	Type       string   `json:"errorType"`
	StackTrace []string `json:"stackTrace"`
}

// Invoke implements the API of the same name from the AWS Lambda API.
// https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html
//
// While the intent is to make this endpoint as similar to the Invoke
// API as possible, there are several features that are not yet
// supported:
//
// -	The "Tail" option for the LogType header does not cause the
//		response to include partial logs.
//
// -	The "Qualifier" parameter is currently ignored and the reported
//		execution version is always "latest".
//
// -	The "Function-Error" header is always "Unhandled" in the event
//		of an exception.
type Invoke struct {
	LogFn      LogFn
	StatFn     StatFn
	URLParamFn URLParamFn
	Fetcher    Fetcher
}

func (h *Invoke) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fnName := h.URLParamFn(r.Context(), "functionName")
	fn, errFn := h.Fetcher.Fetch(r.Context(), fnName)
	switch errFn.(type) {
	case nil:
		break
	case NotFoundError:
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(responseFromError(errFn))
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(responseFromError(errFn))
		return
	}
	fnType := r.Header.Get(invocationTypeHeader)
	if fnType == "" {
		fnType = invocationTypeRequestResponse // This is the default value in AWS.
	}
	ctx := r.Context()
	b, errRead := ioutil.ReadAll(r.Body)
	if errRead != nil {
		w.WriteHeader(http.StatusBadRequest) // Matches JSON parsing errors for the body
		_ = json.NewEncoder(w).Encode(responseFromError(errRead))
		return
	}
	w.Header().Set(invocationVersionHeader, "latest")
	switch fnType {
	case invocationTypeDryRun:
		w.WriteHeader(http.StatusNoContent)
		return
	case invocationTypeEvent:
		ctx = &bgContext{Context: context.Background(), Values: ctx}
		go func() { _, _ = fn.Invoke(ctx, b) }()
		w.WriteHeader(http.StatusAccepted)
	case invocationTypeRequestResponse:
		rb, errInvoke := fn.Invoke(ctx, b)
		statusCode := statusFromError(errInvoke)
		if statusCode > 299 {
			w.Header().Set(invocationErrorHeader, invocationErrorTypeHandled)
		}
		if statusCode > 499 {
			w.Header().Set(invocationErrorHeader, invocationErrorTypeUnhandled)
		}
		w.WriteHeader(statusCode)
		if errInvoke != nil {
			rb, _ = json.Marshal(responseFromError(errInvoke))
		}
		if len(rb) > 0 {
			_, _ = w.Write(rb)
		}
	default:
		w.WriteHeader(http.StatusBadRequest) // Matches the InvalidParameterValueException code
		_ = json.NewEncoder(w).Encode(lambdaError{
			Message:    fmt.Sprintf("InvocationType %s not valid", fnType),
			Type:       "InvalidParameterValueException",
			StackTrace: errResponseStackTrace,
		})
		return
	}
}

// errResponseStackTrace is used to populate the stackTrace attribute of a Lambda
// error. We don't, currently, extract an actual stack trace so we reuse this
// element each time to avoid recreating an empty slice each time.
var errResponseStackTrace = []string{}

func responseFromError(err error) lambdaError {
	errType := reflect.TypeOf(err)
	errTypeName := errType.Name()
	if errType.Kind() == reflect.Ptr {
		errTypeName = errType.Elem().Name()
	}
	return lambdaError{
		Message:    err.Error(),
		Type:       errTypeName,
		StackTrace: errResponseStackTrace,
	}
}

func statusFromError(err error) int {
	switch err.(type) {
	case nil:
		return http.StatusOK
	case *json.InvalidUTF8Error: // nolint
		return http.StatusBadRequest
	case *json.InvalidUnmarshalError:
		return http.StatusBadRequest
	case *json.UnmarshalFieldError: // nolint
		return http.StatusBadRequest
	case *json.UnmarshalTypeError:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
