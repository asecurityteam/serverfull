package domain

import (
	"testing"
)

func TestNotFoundError_Error(t *testing.T) {
	type fields struct {
		ID string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "missing ID",
			fields: fields{},
			want:   "resource () not found",
		},
		{
			name:   "containing ID",
			fields: fields{ID: "test ID"},
			want:   "resource (test ID) not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NotFoundError{
				ID: tt.fields.ID,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("NotFoundError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
