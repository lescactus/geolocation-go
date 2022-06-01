package repositories

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/lescactus/geolocation-go/models"
)

const (
	KeyTTL = 24 * time.Hour
)

type redisDB struct {
	client *redis.Client
	cache  *cache.Cache
}

func NewRedisDB(connstring string) (*redisDB, error) {
	opt, err := redis.ParseURL(connstring)
	if err != nil {
		return nil, fmt.Errorf("error: failed to parse redis url: %w", err)
	}

	client := redis.NewClient(opt)
	cache := cache.New(&cache.Options{
		Redis: client,
	})

	return &redisDB{
		client: client,
		cache:  cache,
	}, nil
}

func (r *redisDB) Save(ctx context.Context, geoip *models.GeoIP) error {
	if err := r.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   geoip.IP,
		Value: geoip,
		TTL:   KeyTTL,
	}); err != nil {
		return fmt.Errorf("error: cannot save value in redis: %w", err)
	}

	return nil
}

func (r *redisDB) Get(ctx context.Context, ip string) (*models.GeoIP, error) {
	var g models.GeoIP
	if err := r.cache.Get(ctx, ip, g); err != nil {
		return nil, fmt.Errorf("error: cannot read value in redis for key: %s: %w", ip, err)
	}

	return &g, nil
}

// Status will retrieve the status of the Redis database.
// It uses the (*redis.Client).Ping() function.
func (r *redisDB) Status(ctx context.Context, wg *sync.WaitGroup, ch chan error) {
	defer wg.Done()
	ch <- r.client.Ping(ctx).Err()
}
