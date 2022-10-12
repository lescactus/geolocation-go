package controllers

import (
	"errors"
	"reflect"
	"testing"
)

func TestHasErrors(t *testing.T) {
	type args struct {
		e map[string]error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "3 items - has 0 error",
			args: args{
				e: map[string]error{
					"one":   nil,
					"two":   nil,
					"three": nil,
				},
			},
			want: false,
		},
		{
			name: "3 items - has 1 error",
			args: args{
				e: map[string]error{
					"one":   errors.New("one"),
					"two":   nil,
					"three": nil,
				},
			},
			want: true,
		},
		{
			name: "3 items - has 2 errors",
			args: args{
				e: map[string]error{
					"one":   errors.New("one"),
					"two":   errors.New("two"),
					"three": nil,
				},
			},
			want: true,
		},
		{
			name: "0 item - has 0 error",
			args: args{
				e: map[string]error{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasErrors(tt.args.e); got != tt.want {
				t.Errorf("hasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewChecks(t *testing.T) {
	type args struct {
		m map[string]error
	}
	tests := []struct {
		name string
		args args
		want []CacheHealthCheck
	}{
		{
			name: "Map is empty - no error",
			args: args{
				m: map[string]error{},
			},
			want: []CacheHealthCheck{},
		},
		{
			name: "Map has one element - one error",
			args: args{
				m: map[string]error{
					"one": errors.New("one"),
				},
			},
			want: []CacheHealthCheck{{CacheName: "one", CacheStatus: HealthzKO, CacheStatusMsg: "one"}},
		},
		{
			name: "Map has one element - no error",
			args: args{
				m: map[string]error{
					"one": nil,
				},
			},
			want: []CacheHealthCheck{{CacheName: "one", CacheStatus: HealthzOK, CacheStatusMsg: "alive"}},
		},
		{
			name: "Map has two elements - two errors",
			args: args{
				m: map[string]error{
					"one": errors.New("one"),
					"two": errors.New("two"),
				},
			},
			want: []CacheHealthCheck{
				{CacheName: "one", CacheStatus: HealthzKO, CacheStatusMsg: "one"},
				{CacheName: "two", CacheStatus: HealthzKO, CacheStatusMsg: "two"},
			},
		},
		{
			name: "Map has two elements - no errors",
			args: args{
				m: map[string]error{
					"one": nil,
					"two": nil,
				},
			},
			want: []CacheHealthCheck{
				{CacheName: "one", CacheStatus: HealthzOK, CacheStatusMsg: "alive"},
				{CacheName: "two", CacheStatus: HealthzOK, CacheStatusMsg: "alive"},
			},
		},
		{
			name: "Map has two elements - one error",
			args: args{
				m: map[string]error{
					"one": nil,
					"two": errors.New("two"),
				},
			},
			want: []CacheHealthCheck{
				{CacheName: "one", CacheStatus: HealthzOK, CacheStatusMsg: "alive"},
				{CacheName: "two", CacheStatus: HealthzKO, CacheStatusMsg: "two"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newChecks(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewChecks() = %v, want %v", got, tt.want)
			}
		})
	}
}
