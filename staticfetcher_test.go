package serverfull

import (
	"context"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestStatic_Fetch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	found := NewMockFunction(ctrl)

	type fields struct {
		Functions map[string]Function
	}
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        Function
		wantErr     bool
		wantErrType reflect.Type
	}{
		{
			name:        "missing function",
			fields:      fields{Functions: make(map[string]Function)},
			args:        args{ctx: context.Background(), name: "missing"},
			want:        nil,
			wantErr:     true,
			wantErrType: reflect.TypeOf(NotFoundError{}),
		},
		{
			name: "found function",
			fields: fields{
				Functions: map[string]Function{
					"found": found,
				},
			},
			args:    args{ctx: context.Background(), name: "found"},
			want:    found,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &StaticFetcher{
				Functions: tt.fields.Functions,
			}
			got, err := f.Fetch(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("StaticFetcher.Fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) && tt.wantErr && tt.wantErrType != nil {
				errType := reflect.TypeOf(err)
				if errType != tt.wantErrType {
					t.Errorf("StaticFetcher.Fetch() error = %v, wantErrType %v", errType, tt.wantErrType)
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StaticFetcher.Fetch() = %v, want %v", got, tt.want)
			}
		})
	}
}
