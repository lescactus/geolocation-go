package controllers

import (
	"github.com/lescactus/geolocation-go/internal/api"
	"github.com/lescactus/geolocation-go/internal/chain"
	"github.com/rs/zerolog"
)

type BaseHandler struct {
	CacheChain  *chain.Chain
	RemoteIPAPI api.GeoAPI
	Logger      *zerolog.Logger
}

func NewBaseHandler(chain *chain.Chain, remoteIPAPI api.GeoAPI, logger *zerolog.Logger) *BaseHandler {
	return &BaseHandler{CacheChain: chain, RemoteIPAPI: remoteIPAPI, Logger: logger}
}
