package handlerfetcher

import (
	"context"
	"reflect"
	"testing"

	"github.com/asecurityteam/serverfull/pkg/domain"
	"github.com/golang/mock/gomock"
)

func TestStatic_FetchHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	found := NewMockHandler(ctrl)

	type fields struct {
		Handlers map[string]domain.Handler
	}
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        domain.Handler
		wantErr     bool
		wantErrType reflect.Type
	}{
		{
			name:        "missing handler",
			fields:      fields{Handlers: make(map[string]domain.Handler)},
			args:        args{ctx: context.Background(), name: "missing"},
			want:        nil,
			wantErr:     true,
			wantErrType: reflect.TypeOf(domain.NotFoundError{}),
		},
		{
			name: "found handler",
			fields: fields{
				Handlers: map[string]domain.Handler{
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
			f := &Static{
				Handlers: tt.fields.Handlers,
			}
			got, err := f.FetchHandler(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Static.FetchHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) && tt.wantErr && tt.wantErrType != nil {
				errType := reflect.TypeOf(err)
				if errType != tt.wantErrType {
					t.Errorf("Static.FetchHandler() error = %v, wantErrType %v", errType, tt.wantErrType)
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Static.FetchHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
