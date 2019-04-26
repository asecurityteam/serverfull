package serverfull

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type mInput struct{}
type mOutput struct{}

func testMFunc(ctx context.Context, in mInput) (mOutput, error) { //nolint
	return mOutput{}, nil
}

func TestMockingFetcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fn := NewMockFunction(ctrl)
	fetcher := NewMockFetcher(ctrl)
	mFetcher := &MockingFetcher{
		Fetcher: fetcher,
	}

	fetcher.EXPECT().Fetch(gomock.Any(), gomock.Any()).Return(fn, nil)
	fn.EXPECT().Source().Return(testMFunc)

	mfn, _ := mFetcher.Fetch(context.Background(), "test")
	require.IsType(t, testMFunc, mfn.Source()) // ensure the mock is the right signature

	res, err := mfn.Invoke(context.Background(), []byte("{}"))
	require.NoError(t, err)
	require.Equal(t, []byte("{}"), res)
}

func TestMockingFetcherError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fetcher := NewMockFetcher(ctrl)
	mFetcher := &MockingFetcher{
		Fetcher: fetcher,
	}

	fetcher.EXPECT().Fetch(gomock.Any(), gomock.Any()).Return(nil, errors.New("fail"))

	_, err := mFetcher.Fetch(context.Background(), "test")
	require.Error(t, err)
}
