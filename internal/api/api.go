package api

import (
	"context"
	"sync"

	"github.com/lescactus/geolocation-go/models"
)

type GeoAPI interface {
	Get(ctx context.Context, ip string) (*models.GeoIP, error)
	Status(ctx context.Context, wg *sync.WaitGroup, ch chan error)
}
