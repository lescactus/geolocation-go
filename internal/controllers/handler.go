package controllers

import (
	"github.com/lescactus/geolocation-go/internal/api"
	"github.com/lescactus/geolocation-go/internal/models"
	"github.com/rs/zerolog"
)

type BaseHandler struct {
	InMemoryRepo models.GeoIPRepository
	RedisRepo    models.GeoIPRepository
	RemoteIPAPI  api.GeoAPI
	Logger       *zerolog.Logger
}

func NewBaseHandler(inMemoryRepo, redisRepo models.GeoIPRepository, remoteIPAPI api.GeoAPI, logger *zerolog.Logger) *BaseHandler {
	return &BaseHandler{InMemoryRepo: inMemoryRepo, RedisRepo: redisRepo, RemoteIPAPI: remoteIPAPI, Logger: logger}
}
