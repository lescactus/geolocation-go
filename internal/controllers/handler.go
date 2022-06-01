package controllers

import (
	"github.com/lescactus/geolocation-go/internal/api"
	"github.com/lescactus/geolocation-go/internal/models"
)

type BaseHandler struct {
	InMemoryRepo models.GeoIPRepository
	RedisRepo    models.GeoIPRepository
	RemoteIPAPI  api.GeoAPI
}

func NewBaseHandler(inMemoryRepo, redisRepo models.GeoIPRepository, remoteIPAPI api.GeoAPI) *BaseHandler {
	return &BaseHandler{InMemoryRepo: inMemoryRepo, RedisRepo: redisRepo, RemoteIPAPI: remoteIPAPI}
}
