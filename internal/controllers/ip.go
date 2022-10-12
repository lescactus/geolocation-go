package controllers

import (
	"context"
	"encoding/json"
	"net"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/lescactus/geolocation-go/internal/models"
	"github.com/rs/zerolog/hlog"
)

const (
	// ContentTypeApplicationJSON represent the applcation/json Content-Type value
	ContentTypeApplicationJSON = "application/json"
)

// GetGeoIP is the main handler.
// It will parse the route variable to ensure it is a valid IPv4 address
// before getting the GeoIP information for the given address.
// It will take care of updating the caches if necessary.
func (h *BaseHandler) GetGeoIP(w http.ResponseWriter, r *http.Request) {
	// Get request id for logging purposes
	req_id, _ := hlog.IDFromCtx(r.Context())

	var ctx = r.Context()

	// Get ip from URL and parse it to a net.IP
	params := httprouter.ParamsFromContext(ctx)
	ip := params.ByName("ip")
	if !isIpv4(ip) {
		h.Logger.Error().Str("req_id", req_id.String()).Msg("the provided IP is not a valid IPv4 address")

		e := NewErrorResponse("the provided ip is not a valid ipv4 address")
		resp, _ := json.Marshal(e)
		w.Header().Set("Content-Type", ContentTypeApplicationJSON)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)

		return
	}

	var g *models.GeoIP
	var err error

	// Lookup in the cache chain for the GeoIP matching the provided ip
	g, err = h.CacheChain.Get(ctx, ip)
	if err != nil {
		h.Logger.Debug().Str("req_id", req_id.String()).Msgf("cache miss from the cache chain: %s", err.Error())

		// Query the remote GeoIP API to retrieve IP information
		g, err = h.RemoteIPAPI.Get(ctx, ip)
		if err != nil {
			h.Logger.Debug().Str("req_id", req_id.String()).Msg("couldn't retrieve geo IP information")

			e := NewErrorResponse("couldn't get geo ip information")
			resp, _ := json.Marshal(e)
			w.Header().Set("Content-Type", ContentTypeApplicationJSON)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(resp)

			return
		}

		// Update all the caches from the chain with the GeoIP
		// Make a new context to be used in the cache save method
		savectx := hlog.CtxWithID(context.Background(), req_id)
		go h.CacheChain.SaveInAllCaches(savectx, g)
	}

	// Marshal the response in json format
	resp, err := json.Marshal(g)
	if err != nil {
		h.Logger.Error().Str("req_id", req_id.String()).Msg("couldn't marshal geo ip information")
		e := NewErrorResponse("couldn't marshal geo ip information")
		resp, _ := json.Marshal(e)
		w.Header().Set("Content-Type", ContentTypeApplicationJSON)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resp)
	}

	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// isIpv4 verify the given string is a valid IPv4 address.
// Return true if yes, false otherwise
func isIpv4(host string) bool {
	return net.ParseIP(host) != nil
}
