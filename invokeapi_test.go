package serverfull

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type nopLogger struct{}

func (*nopLogger) Debug(event interface{})                 {}
func (*nopLogger) Info(event interface{})                  {}
func (*nopLogger) Warn(event interface{})                  {}
func (*nopLogger) Error(event interface{})                 {}
func (*nopLogger) SetField(name string, value interface{}) {}
func (logger *nopLogger) Copy() Logger {
	return logger
}

var testLogger = &nopLogger{}

func testLogFn(context.Context) Logger { return testLogger }

type nopStat struct{}

func (*nopStat) Gauge(stat string, value float64, tags ...string)        {}
func (*nopStat) Count(stat string, count float64, tags ...string)        {}
func (*nopStat) Histogram(stat string, value float64, tags ...string)    {}
func (*nopStat) Timing(stat string, value time.Duration, tags ...string) {}
func (*nopStat) AddTags(tags ...string)                                  {}
func (*nopStat) GetTags() []string {
	return []string{}
}

var testStat = &nopStat{}

func testStatFn(context.Context) Stat { return testStat }

var testName = "test"

type testCtxKey string

var (
	ctxKey  testCtxKey = "key"
	ctxKey2 testCtxKey = "key2"
)

type URLParam string

func (p URLParam) Get(context.Context, string) string {
	return string(p)
}

func TestBackgroundContext(t *testing.T) {
	original, cancelOriginal := context.WithCancel(context.Background())
	original = context.WithValue(original, ctxKey, "value")
	defer cancelOriginal()

	var bg context.Context = &bgContext{
		Context: context.Background(),
		Values:  original,
	}
	bg = context.WithValue(bg, ctxKey2, "value2")
	bg, cancelBg := context.WithCancel(bg)
	defer cancelBg()

	v := bg.Value(ctxKey)
	assert.IsType(t, "", v, "bgContext did not preserve values")
	assert.Equal(t, v, "value")
	v = bg.Value(ctxKey2)
	assert.IsType(t, "", v, "bgContext did not expose new values")
	assert.Equal(t, v, "value2")

	cancelOriginal()
	select {
	case <-bg.Done():
		assert.Fail(t, "bgContext was prematurely canceled")
	default:
	}

	cancelBg()
	select {
	case <-bg.Done():
	default:
		assert.Fail(t, "bgContext did respect it's own cancelation")
	}
}

func Test_statusFromError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "nil",
			args: args{err: nil},
			want: http.StatusOK,
		},
		{
			name: "*json.InvalidUTF8Error",
			args: args{err: &json.InvalidUTF8Error{}}, // nolint
			want: http.StatusBadRequest,
		},
		{
			name: "*json.InvalidUnmarshalError",
			args: args{err: &json.InvalidUnmarshalError{}},
			want: http.StatusBadRequest,
		},
		{
			name: "*json.UnmarshalFieldError",
			args: args{err: &json.UnmarshalFieldError{}}, // nolint
			want: http.StatusBadRequest,
		},
		{
			name: "*json.UnmarshalTypeError",
			args: args{err: &json.UnmarshalTypeError{}},
			want: http.StatusBadRequest,
		},
		{
			name: "unknown",
			args: args{err: errors.New(testName)},
			want: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusFromError(tt.args.err); got != tt.want {
				t.Errorf("statusFromError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_responseFromError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want lambdaError
	}{
		{
			name: "non-pointer",
			args: args{err: NotFoundError{ID: testName}},
			want: lambdaError{
				Message:    NotFoundError{ID: testName}.Error(),
				Type:       "NotFoundError",
				StackTrace: errResponseStackTrace,
			},
		},
		{
			name: "pointer",
			args: args{err: &NotFoundError{ID: testName}},
			want: lambdaError{
				Message:    NotFoundError{ID: testName}.Error(),
				Type:       "NotFoundError",
				StackTrace: errResponseStackTrace,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := responseFromError(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("responseFromError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInvokeFunctionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fnName := testName
	fetcher := NewMockFetcher(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      testLogFn,
		StatFn:     testStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	r, _ := http.NewRequest(http.MethodPost, path, http.NoBody)

	fetcher.EXPECT().Fetch(gomock.Any(), fnName).Return(nil, NotFoundError{ID: fnName})
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestInvokeFunctionFetchFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fnName := testName
	fetcher := NewMockFetcher(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      testLogFn,
		StatFn:     testStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	r, _ := http.NewRequest(http.MethodPost, path, http.NoBody)

	fetcher.EXPECT().Fetch(gomock.Any(), fnName).Return(nil, errors.New("fail"))
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestInvokeFunctionInvalidInvocationType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fnName := testName
	fetcher := NewMockFetcher(ctrl)
	fn := NewMockFunction(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      testLogFn,
		StatFn:     testStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	r, _ := http.NewRequest(http.MethodPost, path, http.NoBody)
	r.Header.Set(invocationTypeHeader, "unknown")

	fetcher.EXPECT().Fetch(gomock.Any(), fnName).Return(fn, nil)
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInvokeFunctionDryRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fnName := testName
	fetcher := NewMockFetcher(ctrl)
	fn := NewMockFunction(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      testLogFn,
		StatFn:     testStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	r, _ := http.NewRequest(http.MethodPost, path, http.NoBody)
	r.Header.Set(invocationTypeHeader, invocationTypeDryRun)

	fetcher.EXPECT().Fetch(gomock.Any(), fnName).Return(fn, nil)
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestInvokeFunctionEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	done := make(chan interface{})
	fnName := testName
	fetcher := NewMockFetcher(ctrl)
	fn := NewMockFunction(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      testLogFn,
		StatFn:     testStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	input := []byte("data")
	output := []byte("response")
	r, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(input))
	r.Header.Set(invocationTypeHeader, invocationTypeEvent)

	fetcher.EXPECT().Fetch(gomock.Any(), fnName).Return(fn, nil)
	fn.EXPECT().Invoke(gomock.Any(), input).Do(func(context.Context, []byte) {
		close(done)
	}).Return(output, nil)
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusAccepted, w.Code)
	select {
	case <-done:
	case <-time.After(time.Second):
		assert.Fail(t, "event was not executed in the background")
	}
}

func TestInvokeFunctionRequestResponseBadInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fnName := testName
	fetcher := NewMockFetcher(ctrl)
	fn := NewMockFunction(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      testLogFn,
		StatFn:     testStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	input := []byte("data")
	r, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(input))

	fetcher.EXPECT().Fetch(gomock.Any(), fnName).Return(fn, nil)
	fn.EXPECT().Invoke(gomock.Any(), input).Return(nil, &json.InvalidUnmarshalError{Type: reflect.TypeOf(1)})
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInvokeFunctionRequestResponseFunctionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fnName := testName
	fetcher := NewMockFetcher(ctrl)
	fn := NewMockFunction(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      testLogFn,
		StatFn:     testStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	input := []byte("data")
	r, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(input))

	fetcher.EXPECT().Fetch(gomock.Any(), fnName).Return(fn, nil)
	fn.EXPECT().Invoke(gomock.Any(), input).Return(nil, errors.New("fail"))
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestInvokeFunctionRequestResponseSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fnName := testName
	fetcher := NewMockFetcher(ctrl)
	fn := NewMockFunction(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      testLogFn,
		StatFn:     testStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	input := []byte("data")
	output := []byte("response")
	r, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(input))

	fetcher.EXPECT().Fetch(gomock.Any(), fnName).Return(fn, nil)
	fn.EXPECT().Invoke(gomock.Any(), input).Return(output, nil)
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, output, w.Body.Bytes())
}

func TestInvokeErrorNoMockMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fnName := testName
	fetcher := NewMockFetcher(ctrl)
	fn := NewMockFunction(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      testLogFn,
		StatFn:     testStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	input := []byte("data")
	r, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(input))
	r.Header.Set(invocationTypeHeader, invocationTypeError)

	fetcher.EXPECT().Fetch(gomock.Any(), fnName).Return(fn, nil)
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInvokeErrorNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fnName := testName
	fetcher := NewMockFetcher(ctrl)
	fn := NewMockFunction(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      testLogFn,
		StatFn:     testStatFn,
		URLParamFn: URLParam(fnName).Get,
		MockMode:   true,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	input := []byte("data")
	r, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(input))
	r.Header.Set(invocationTypeHeader, invocationTypeError)
	r.Header.Set(invocationErrorTypeHeader, "notexists")

	fetcher.EXPECT().Fetch(gomock.Any(), fnName).Return(fn, nil)
	fn.EXPECT().Errors().Return(nil)
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestInvokeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErr := errors.New("testError")
	fnName := testName
	fetcher := NewMockFetcher(ctrl)
	fn := NewMockFunction(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      testLogFn,
		StatFn:     testStatFn,
		URLParamFn: URLParam(fnName).Get,
		MockMode:   true,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	input := []byte("data")
	r, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(input))
	r.Header.Set(invocationTypeHeader, invocationTypeError)
	r.Header.Set(invocationErrorTypeHeader, "errorString")

	fetcher.EXPECT().Fetch(gomock.Any(), fnName).Return(fn, nil)
	fn.EXPECT().Errors().Return([]error{mockErr})
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
