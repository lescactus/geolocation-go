package chain

import (
	"context"
	"errors"
	"sync"

	"github.com/lescactus/geolocation-go/internal/models"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

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
	geoIPRepository map[string]models.GeoIPRepository
	l               *zerolog.Logger
}

// New will return a new empty chain.
// It is then up to the caller to use the Add()
// method to add GeoIPRepository to the chain.
func New(l *zerolog.Logger) *Chain {
	m := make(map[string]models.GeoIPRepository)
	c := new(Chain)
	c.geoIPRepository = m
	c.l = l
	return c
}

// Add will add a models.GeoIPRepository to the chain.
// Return an error if the GeoIPRepository is already present.
func (c *Chain) Add(name string, g models.GeoIPRepository) error {
	if _, ok := c.geoIPRepository[name]; ok {
		return errors.New("error: GeoIPRepository already present in chain")
	}

	c.geoIPRepository[name] = g

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

	for name, repo := range c.geoIPRepository {
		c.l.Trace().Str("req_id", req_id).Msgf("looking for %s in %s database from cache chain", ip, name)

		g, err = repo.Get(ctx, ip)
		if err != nil {
			c.l.Debug().Str("req_id", req_id).Err(err).Msgf("cache miss from %s database", name)
			cacheMiss = append(cacheMiss, name)
		} else {
			c.l.Debug().Str("req_id", req_id).Msgf("cache hit from %s database", name)

			// Update all caches with g if needed
			if len(cacheMiss) > 0 {
				go c.saveInCaches(ctx, g, cacheMiss)
			}
			return g, nil
		}
	}

	return nil, errors.New("couldn't find entry in cache chain")
}

// Statuses will call each GeoIPRepository Status() function concurrently and return errors
// if any.
func (c *Chain) Statuses(ctx context.Context) map[string]error {
	var wg sync.WaitGroup

	errors := make(map[string]error)

	for name, repo := range c.geoIPRepository {
		c := make(chan error, 1)
		wg.Add(1)
		go repo.Status(ctx, &wg, c)

		errors[name] = <-c
	}
	wg.Wait()

	return errors
}

// SaveInAllCaches will save geoip asynchronousely in all the caches from the chain.
func (c *Chain) SaveInAllCaches(ctx context.Context, geoip *models.GeoIP) {
	var caches = make([]string, 0, len(c.geoIPRepository))
	for name := range c.geoIPRepository {
		caches = append(caches, name)
	}

	req_id := reqIDFromContext(ctx)
	c.l.Trace().Str("req_id", req_id).Msgf("preparing to update following caches: %+v", caches)

	c.saveInCaches(ctx, geoip, caches)
}

// saveInCaches will save geoip asynchronousely in the given cache(s).
func (c *Chain) saveInCaches(ctx context.Context, geoip *models.GeoIP, caches []string) {
	var wg sync.WaitGroup

	req_id := reqIDFromContext(ctx)

	wg.Add(len(caches))

	for _, name := range caches {
		go func(name string) {
			defer wg.Done()

			c.l.Debug().Str("req_id", req_id).Msgf("updating cache %s with entry %s", name, geoip.IP)
			if err := c.geoIPRepository[name].Save(ctx, geoip); err != nil {
				c.l.Error().Str("req_id", req_id).Msgf("fail to cache in %s database: %s", name, err.Error())
			} else {
				c.l.Trace().Str("req_id", req_id).Msgf("cache %s updated with entry %s", name, geoip.IP)
			}

		}(name)
	}

	wg.Wait()
}

// GeoIPRepositoryLen returns the lenght of the geoIPRepository map
func (c *Chain) GeoIPRepositoryLen() int {
	return len(c.geoIPRepository)
}

// reqIDFromContext extracts and returns the request id from
// the given context.
func reqIDFromContext(ctx context.Context) string {
	req_id, _ := hlog.IDFromCtx(ctx)
	return req_id.String()
}
