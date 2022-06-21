package repositories

import (
	"context"
	"fmt"
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
		ttl        time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantTTL time.Duration
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
			wantTTL: DefaultKeyTTL,
			wantErr: false,
		},
		{
			name:    "Valid connection string - TTL 1h - redis://localhost:6379",
			args:    args{connstring: "redis://localhost:6379", ttl: time.Hour},
			wantTTL: time.Hour,
			wantErr: false,
		},
		{
			name:    "Valid connection string - redis://user:pass@localhost:6379/1",
			args:    args{connstring: "redis://user:pass@localhost:6379/1"},
			wantTTL: DefaultKeyTTL,
			wantErr: false,
		},
		{
			name:    "Valid connection string - redis://user:pass@localhost:6379/1?dial_timeout=3&db=1&read_timeout=6s&max_retries=2",
			args:    args{connstring: "redis://user:pass@localhost:6379/1?dial_timeout=3&db=1&read_timeout=6s&max_retries=2"},
			wantTTL: DefaultKeyTTL,
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
			r, err := NewRedisDB(tt.args.connstring, tt.args.ttl)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRedisDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (err == nil) && (tt.wantTTL != r.keyTTL) {
				t.Errorf("NewRedisDB() error = %v, wantTTL %v", err, tt.wantTTL)
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
		keyTTL time.Duration
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
			fields:  fields{client: redis.NewClient(&redis.Options{Addr: s.Addr()}), cache: cache.New(&cache.Options{Redis: client}), keyTTL: time.Hour},
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
			assert.Equal(t, tt.fields.keyTTL, s.TTL(tt.args.geoip.IP))

			if tt.wantErr == false {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func BenchmarkRedisDBSave(b *testing.B) {
	s := miniredis.RunT(b)
	//client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	r, err := NewRedisDB(fmt.Sprintf("redis://%s", s.Addr()), time.Hour)
	if err != nil {
		b.Error(err)
	}

	var ctx = context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Save(ctx, &models.GeoIP{IP: "1.1.1.1"})
	}
}
