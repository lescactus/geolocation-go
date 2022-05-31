package api

import (
	"context"

	"github.com/lescactus/geolocation-go/models"
)

type GeoAPI interface {
	Get(ctx context.Context, ip string) (*models.GeoIP, error)
}
