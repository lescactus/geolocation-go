package chain

import (
	"context"
	"errors"
	"sync"

	"github.com/lescactus/geolocation-go/internal/models"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// Cache is a models.GeoIPRepository and is used within a
// Chain.
type Cache struct {
	name       string
	repository models.GeoIPRepository
}

// Chain acts as a list of GeoIPRepository.
// It's goal is to provide a cache to read and write models.GeoIP
// in a chained way: when requesting a read, the first GeoIPRepository
// will be queried and in case of cache miss, the second GeoIPRepository will be queried,
// then in case of cache miss, the third GeoIPRepository will be queried and so on...
// When one of the cache is a hit, all the other caches will be updated if necessary.
//
//
// Typically, the chain would be composed of several caches stores, starting from the fastest to
// the slowest.
//
// Ex:
//
// 1. In memory cache - very fast
//    |
//    V
// 2. Redis - fast but slower than in memory cache
//    |
//    V
// 3. MySQL - fast but slower than redis
//
type Chain struct {
	caches []Cache
	l      *zerolog.Logger
}

// New will return a new empty chain.
// It is then up to the caller to use the Add()
// method to add GeoIPRepository to the chain.
func New(l *zerolog.Logger) *Chain {
	return &Chain{
		l:      l,
		caches: make([]Cache, 0),
	}
}

// Add will add a models.GeoIPRepository to the chain.
// Return an error if the GeoIPRepository is already present.
func (c *Chain) Add(name string, g models.GeoIPRepository) error {
	if g == nil {
		return errors.New("error: GeoIPRepository cannot be nil")
	}

	if len(c.caches) == 0 {
		c.caches = append(c.caches, Cache{
			name:       name,
			repository: g,
		})
		return nil
	}

	for _, cache := range c.caches {
		if cache.name == name {
			return errors.New("error: GeoIPRepository already present in chain")
		}
	}

	c.caches = append(c.caches, Cache{name: name, repository: g})

	return nil
}

// Get will attempt to retrieve the *models.GeoIP corresponding to the given ip.
//
// It will lookup each GeoIPRepository in the chain and return the first *models.GeoIP
// found.
// Each GeoIPRepository which doesn't have the *models.GeoIP will be updated accordingly.
//
// If no *models.GeoIP is found, an error will be returned.
func (c *Chain) Get(ctx context.Context, ip string) (*models.GeoIP, error) {
	var g *models.GeoIP
	var err error
	var cacheMiss []string

	req_id := reqIDFromContext(ctx)

	for _, cache := range c.caches {
		c.l.Trace().Str("req_id", req_id).Msgf("looking for %s in %s database from cache chain", ip, cache.name)

		g, err = cache.repository.Get(ctx, ip)
		if err != nil {
			c.l.Debug().Str("req_id", req_id).Err(err).Msgf("cache miss from %s database", cache.name)
			cacheMiss = append(cacheMiss, cache.name)
		} else {
			c.l.Debug().Str("req_id", req_id).Msgf("cache hit from %s database", cache.name)

			// Update all caches with g if needed
			if len(cacheMiss) > 0 {
				go c.SaveInAllCaches(ctx, g)
			}
			return g, nil
		}
	}

	return nil, errors.New("couldn't find entry in cache chain")
}

// Statuses will call each GeoIPRepository Status() function concurrently
// and return errors if any.
func (c *Chain) Statuses(ctx context.Context) map[string]error {
	var wg sync.WaitGroup

	errors := make(map[string]error)

	for _, cache := range c.caches {
		c := make(chan error, 1)
		wg.Add(1)
		go cache.repository.Status(ctx, &wg, c)

		errors[cache.name] = <-c
	}
	wg.Wait()

	return errors
}

// SaveInAllCaches will save geoip asynchronousely in all the caches from the chain.
func (c *Chain) SaveInAllCaches(ctx context.Context, geoip *models.GeoIP) {
	req_id := reqIDFromContext(ctx)

	var wg sync.WaitGroup
	wg.Add(len(c.caches))

	for _, cache := range c.caches {
		go func(cache Cache) {
			defer wg.Done()

			c.l.Debug().Str("req_id", req_id).Msgf("updating cache %s with entry %s", cache.name, geoip.IP)
			if err := cache.repository.Save(ctx, geoip); err != nil {
				c.l.Error().Str("req_id", req_id).Msgf("fail to cache in %s database: %s", cache.name, err.Error())
			} else {
				c.l.Trace().Str("req_id", req_id).Msgf("cache %s updated with entry %s", cache.name, geoip.IP)
			}
		}(cache)
	}

	wg.Wait()
}

// reqIDFromContext extracts and returns the request id from
// the given context.
func reqIDFromContext(ctx context.Context) string {
	req_id, _ := hlog.IDFromCtx(ctx)
	return req_id.String()
}
