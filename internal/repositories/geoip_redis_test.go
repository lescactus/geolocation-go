package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/lescactus/geolocation-go/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisDB(t *testing.T) {
	type args struct {
		connstring string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Empty connection string",
			args:    args{connstring: ""},
			wantErr: true,
		},
		{
			name:    "Valid connection string - redis://localhost:6379",
			args:    args{connstring: "redis://localhost:6379"},
			wantErr: false,
		},
		{
			name:    "Valid connection string - redis://user:pass@localhost:6379/1",
			args:    args{connstring: "redis://user:pass@localhost:6379/1"},
			wantErr: false,
		},
		{
			name:    "Valid connection string - redis://user:pass@localhost:6379/1?dial_timeout=3&db=1&read_timeout=6s&max_retries=2",
			args:    args{connstring: "redis://user:pass@localhost:6379/1?dial_timeout=3&db=1&read_timeout=6s&max_retries=2"},
			wantErr: false,
		},
		{
			name:    "Invalid connection string - redis://user:pass@localhost:6379/db",
			args:    args{connstring: "redis://user:pass@localhost:6379/db"},
			wantErr: true,
		},
		{
			name:    "Invalid connection string - azertyuiop",
			args:    args{connstring: "azertyuiop"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRedisDB(tt.args.connstring)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRedisDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRedisDBSave(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})

	type fields struct {
		client *redis.Client
		cache  *cache.Cache
	}
	type args struct {
		ctx   context.Context
		geoip *models.GeoIP
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Save ",
			fields:  fields{client: redis.NewClient(&redis.Options{Addr: s.Addr()}), cache: cache.New(&cache.Options{Redis: client})},
			args:    args{ctx: context.Background(), geoip: &models.GeoIP{IP: "1.1.1.1"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &redisDB{
				client: tt.fields.client,
				cache:  tt.fields.cache,
			}

			err := r.Save(tt.args.ctx, tt.args.geoip)
			assert.True(t, s.Exists(tt.args.geoip.IP))
			assert.Equal(t, 24*time.Hour, s.TTL(tt.args.geoip.IP))

			if tt.wantErr == false {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
