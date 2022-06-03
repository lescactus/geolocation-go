package controllers

import (
	"reflect"
	"testing"
)

func TestNewErrorResponse(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
		want *ErrorResponse
	}{
		{
			name: "Non empty message",
			args: args{"this is an error"},
			want: &ErrorResponse{Status: "error", Msg: "this is an error"},
		},
		{
			name: "Empty message",
			args: args{""},
			want: &ErrorResponse{Status: "error", Msg: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewErrorResponse(tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewErrorResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
