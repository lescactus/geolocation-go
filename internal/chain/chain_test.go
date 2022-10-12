package chain

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/lescactus/geolocation-go/internal/models"
	"github.com/lescactus/geolocation-go/internal/repositories"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func TestNew(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	type args struct {
		l *zerolog.Logger
	}
	tests := []struct {
		name string
		args args
		want *Chain
	}{
		{
			name: "With logger",
			args: args{&logger},
			want: &Chain{geoIPRepository: make(map[string]models.GeoIPRepository), l: &logger},
		},
		{
			name: "With nil logger",
			args: args{nil},
			want: &Chain{geoIPRepository: make(map[string]models.GeoIPRepository), l: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChainAdd(t *testing.T) {
	logger := zerolog.New(os.Stdout)

	m := make(map[string]models.GeoIPRepository)
	m["in-memory"] = repositories.NewInMemoryDB()

	type fields struct {
		geoIPRepository map[string]models.GeoIPRepository
		l               *zerolog.Logger
	}
	type args struct {
		name string
		g    models.GeoIPRepository
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Add nil",
			fields:  fields{make(map[string]models.GeoIPRepository), &logger},
			args:    args{"", nil},
			wantErr: false,
		},
		{
			name:    "Add in-memory",
			fields:  fields{make(map[string]models.GeoIPRepository), &logger},
			args:    args{"in-memory", repositories.NewInMemoryDB()},
			wantErr: false,
		},
		{
			name:    "Add already existing GeoIPRepository",
			fields:  fields{m, &logger},
			args:    args{"in-memory", repositories.NewInMemoryDB()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Chain{
				geoIPRepository: tt.fields.geoIPRepository,
				l:               tt.fields.l,
			}
			if err := c.Add(tt.args.name, tt.args.g); (err != nil) != tt.wantErr {
				t.Errorf("Chain.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReqIDFromContext(t *testing.T) {
	var ctxWithID context.Context
	var ctxWithoutID context.Context

	ctxWithID = hlog.CtxWithID(context.Background(), xid.New())
	ctxWithoutID = context.Background()
	req_id, _ := hlog.IDFromCtx(ctxWithID)

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Request ID is present in context",
			args: args{ctxWithID},
			want: req_id.String(),
		},
		{
			name: "Request ID is not present in context",
			args: args{ctxWithoutID},
			want: "00000000000000000000", // default value when id isn't set
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reqIDFromContext(tt.args.ctx); got != tt.want {
				t.Errorf("reqIDFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChainGeoIPRepositoryLen(t *testing.T) {
	logger := zerolog.New(os.Stdout)

	m := make(map[string]models.GeoIPRepository)
	m["in-memory"] = repositories.NewInMemoryDB()

	type fields struct {
		geoIPRepository map[string]models.GeoIPRepository
		l               *zerolog.Logger
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "Without GeoIPRepository",
			fields: fields{make(map[string]models.GeoIPRepository), &logger},
			want:   0,
		},
		{
			name:   "With one GeoIPRepository",
			fields: fields{m, &logger},
			want:   1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Chain{
				geoIPRepository: tt.fields.geoIPRepository,
				l:               tt.fields.l,
			}
			if got := c.GeoIPRepositoryLen(); got != tt.want {
				t.Errorf("Chain.GeoIPRepositoryLen() = %v, want %v", got, tt.want)
			}
		})
	}
}
