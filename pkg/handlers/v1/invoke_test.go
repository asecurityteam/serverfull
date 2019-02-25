package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/asecurityteam/logevent"
	"github.com/asecurityteam/serverfull/pkg/domain"
	"github.com/golang/mock/gomock"
	"github.com/rs/xstats"
	"github.com/stretchr/testify/assert"
)

var (
	nullLogger = logevent.New(logevent.Config{Output: ioutil.Discard})
	nullLogFn  = func(context.Context) domain.Logger { return nullLogger }
	nullStatFn = xstats.FromContext
)

type URLParam string

func (p URLParam) Get(context.Context, string) string {
	return string(p)
}

func TestBackgroundContext(t *testing.T) {
	original, cancelOriginal := context.WithCancel(context.Background())
	original = context.WithValue(original, "key", "value")
	defer cancelOriginal()

	var bg context.Context = &bgContext{
		Context: context.Background(),
		Values:  original,
	}
	bg = context.WithValue(bg, "key2", "value2")
	bg, cancelBg := context.WithCancel(bg)
	defer cancelBg()

	v := bg.Value("key")
	assert.IsType(t, "", v, "bgContext did not preserve values")
	assert.Equal(t, v, "value")
	v = bg.Value("key2")
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
			args: args{err: &json.InvalidUTF8Error{}},
			want: http.StatusBadRequest,
		},
		{
			name: "*json.InvalidUnmarshalError",
			args: args{err: &json.InvalidUnmarshalError{}},
			want: http.StatusBadRequest,
		},
		{
			name: "*json.UnmarshalFieldError",
			args: args{err: &json.UnmarshalFieldError{}},
			want: http.StatusBadRequest,
		},
		{
			name: "*json.UnmarshalTypeError",
			args: args{err: &json.UnmarshalTypeError{}},
			want: http.StatusBadRequest,
		},
		{
			name: "unknown",
			args: args{err: errors.New("test")},
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
			args: args{err: domain.NotFoundError{ID: "test"}},
			want: lambdaError{
				Message:    domain.NotFoundError{ID: "test"}.Error(),
				Type:       "NotFoundError",
				StackTrace: errResponseStackTrace,
			},
		},
		{
			name: "pointer",
			args: args{err: &domain.NotFoundError{ID: "test"}},
			want: lambdaError{
				Message:    domain.NotFoundError{ID: "test"}.Error(),
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

	fnName := "test"
	fetcher := NewMockHandlerFetcher(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      nullLogFn,
		StatFn:     nullStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	r, _ := http.NewRequest(http.MethodPost, path, http.NoBody)

	fetcher.EXPECT().FetchHandler(gomock.Any(), fnName).Return(nil, domain.NotFoundError{ID: fnName})
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestInvokeFunctionFetchFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fnName := "test"
	fetcher := NewMockHandlerFetcher(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      nullLogFn,
		StatFn:     nullStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	r, _ := http.NewRequest(http.MethodPost, path, http.NoBody)

	fetcher.EXPECT().FetchHandler(gomock.Any(), fnName).Return(nil, errors.New("fail"))
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestInvokeFunctionInvalidInvocationType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fnName := "test"
	fetcher := NewMockHandlerFetcher(ctrl)
	fn := NewMockHandler(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      nullLogFn,
		StatFn:     nullStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	r, _ := http.NewRequest(http.MethodPost, path, http.NoBody)
	r.Header.Set(invocationTypeHeader, "unknown")

	fetcher.EXPECT().FetchHandler(gomock.Any(), fnName).Return(fn, nil)
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInvokeFunctionDryRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fnName := "test"
	fetcher := NewMockHandlerFetcher(ctrl)
	fn := NewMockHandler(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      nullLogFn,
		StatFn:     nullStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	r, _ := http.NewRequest(http.MethodPost, path, http.NoBody)
	r.Header.Set(invocationTypeHeader, invocationTypeDryRun)

	fetcher.EXPECT().FetchHandler(gomock.Any(), fnName).Return(fn, nil)
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestInvokeFunctionEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	done := make(chan interface{})
	fnName := "test"
	fetcher := NewMockHandlerFetcher(ctrl)
	fn := NewMockHandler(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      nullLogFn,
		StatFn:     nullStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	input := []byte("data")
	output := []byte("response")
	r, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(input))
	r.Header.Set(invocationTypeHeader, invocationTypeEvent)

	fetcher.EXPECT().FetchHandler(gomock.Any(), fnName).Return(fn, nil)
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

	fnName := "test"
	fetcher := NewMockHandlerFetcher(ctrl)
	fn := NewMockHandler(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      nullLogFn,
		StatFn:     nullStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	input := []byte("data")
	r, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(input))

	fetcher.EXPECT().FetchHandler(gomock.Any(), fnName).Return(fn, nil)
	fn.EXPECT().Invoke(gomock.Any(), input).Return(nil, &json.InvalidUnmarshalError{Type: reflect.TypeOf(1)})
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInvokeFunctionRequestResponseFunctionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fnName := "test"
	fetcher := NewMockHandlerFetcher(ctrl)
	fn := NewMockHandler(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      nullLogFn,
		StatFn:     nullStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	input := []byte("data")
	r, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(input))

	fetcher.EXPECT().FetchHandler(gomock.Any(), fnName).Return(fn, nil)
	fn.EXPECT().Invoke(gomock.Any(), input).Return(nil, errors.New("fail"))
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestInvokeFunctionRequestResponseSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fnName := "test"
	fetcher := NewMockHandlerFetcher(ctrl)
	fn := NewMockHandler(ctrl)
	handler := &Invoke{
		Fetcher:    fetcher,
		LogFn:      nullLogFn,
		StatFn:     nullStatFn,
		URLParamFn: URLParam(fnName).Get,
	}
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/2015-03-31/functions/%s/invocations", fnName)
	input := []byte("data")
	output := []byte("response")
	r, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(input))

	fetcher.EXPECT().FetchHandler(gomock.Any(), fnName).Return(fn, nil)
	fn.EXPECT().Invoke(gomock.Any(), input).Return(output, nil)
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, output, w.Body.Bytes())
}
