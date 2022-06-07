package repositories

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/lescactus/geolocation-go/internal/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	KeyTTL = 24 * time.Hour
)

// Prometheus metrics
var (
	redisItemSaved = promauto.NewCounter(prometheus.CounterOpts{
		Name: "redis_items_saved_total",
		Help: "The total number of saved items in the redis database",
	})
	redisItemFailedSaved = promauto.NewCounter(prometheus.CounterOpts{
		Name: "redis_items_failed_saved_total",
		Help: "The total number of failed saved items in the redis database",
	})
	redisItemRead = promauto.NewCounter(prometheus.CounterOpts{
		Name: "redis_items_read_total",
		Help: "The total number of read items from the redis database",
	})
	redisItemFailedRead = promauto.NewCounter(prometheus.CounterOpts{
		Name: "redis_items_failed_read_total",
		Help: "The total number of failed read items from the redis database",
	})
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
		// Increment Prometheus counter
		redisItemFailedSaved.Inc()
		return fmt.Errorf("error: cannot save value in redis: %w", err)
	}

	// Increment the Prometheus counter
	redisItemSaved.Inc()

	return nil
}

func (r *redisDB) Get(ctx context.Context, ip string) (*models.GeoIP, error) {
	var g models.GeoIP
	if err := r.cache.Get(ctx, ip, &g); err != nil {
		// Increment Prometheus counter
		redisItemFailedRead.Inc()
		return nil, fmt.Errorf("error: cannot read value in redis for key: %s: %w", ip, err)
	}

	// Increment the Prometheus counter
	redisItemRead.Inc()

	return &g, nil
}

// Status will retrieve the status of the Redis database.
// It uses the (*redis.Client).Ping() function.
func (r *redisDB) Status(ctx context.Context, wg *sync.WaitGroup, ch chan error) {
	defer wg.Done()
	ch <- r.client.Ping(ctx).Err()
}
