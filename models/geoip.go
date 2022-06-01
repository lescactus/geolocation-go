package models

import (
	"context"
	"sync"
)

// GeoIP contains IP Geolocation information
type GeoIP struct {
	IP          string  `json:"ip"`
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	City        string  `json:"city,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
}

type GeoIPRepository interface {
	Get(ctx context.Context, ip string) (*GeoIP, error)
	Save(ctx context.Context, geoip *GeoIP) error
	Status(ctx context.Context, wg *sync.WaitGroup, ch chan error)
}
