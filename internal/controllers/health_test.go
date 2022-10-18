package controllers

import (
	"errors"
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
			name: "Map has two elements - no error",
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
		{
			name: "Map has three elements - no error",
			args: args{
				m: map[string]error{
					"one":   nil,
					"two":   nil,
					"three": nil,
				},
			},
			want: []CacheHealthCheck{
				{CacheName: "one", CacheStatus: HealthzOK, CacheStatusMsg: "alive"},
				{CacheName: "two", CacheStatus: HealthzOK, CacheStatusMsg: "alive"},
				{CacheName: "three", CacheStatus: HealthzOK, CacheStatusMsg: "alive"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newChecks(tt.args.m)

			if len(got) != len(tt.want) {
				t.Errorf("len(got) = %v, want %v", len(got), len(tt.want))
			}

			for name, err := range tt.args.m {
				var c CacheHealthCheck
				for _, v := range tt.want {
					if name == v.CacheName {
						c = v

						if err != nil {
							if v.CacheStatus != HealthzKO {
								t.Errorf("Invalid cache status for %v: expected %v, got %v", v, HealthzKO, v.CacheStatus)
							}
							if v.CacheStatusMsg != err.Error() {
								t.Errorf("Invalid cache status message for %v: expected %v, got %v", v, err.Error(), v.CacheStatusMsg)
							}
						} else {
							if v.CacheStatus != HealthzOK {
								t.Errorf("Invalid cache status for %v: expected %v, got %v", v, HealthzOK, v.CacheStatus)
							}
							if v.CacheStatusMsg != "alive" {
								t.Errorf("Invalid cache status message for %v: expected %v, got %v", v, "alive", v.CacheStatusMsg)
							}
						}

					}
				}
				if c == (CacheHealthCheck{}) {
					t.Errorf("Missing CacheHealthCheck %v => %v", name, err)
				}
			}
		})
	}
}
