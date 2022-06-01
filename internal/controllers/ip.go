package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lescactus/geolocation-go/internal/models"
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
	// Get ip from URL and parse it to a net.IP
	ip := mux.Vars(r)["ip"]
	if !isIpv4(ip) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`error: the provided IP is not a valid IPv4 address`))
		return
	}

	var ctx = context.Background()
	var g *models.GeoIP
	var err error

	// KEEP IT SIMPLE
	// TODO: Implement custom errors

	// Retrieve the IP information from the in-memory database
	g, err = h.InMemoryRepo.Get(ctx, ip)
	if err != nil {
		log.Println("cache miss from in-memory database")
		// Retrieve the IP information from the redis database
		g, err = h.RedisRepo.Get(ctx, ip)
		if err != nil {
			log.Println("cache miss from redis database")
			// Query the remote GeoIP API to retrieve IP information
			g, err = h.RemoteIPAPI.Get(ctx, ip)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`error: couldn't retrieve geo IP information`))
				return
			}

			// Save the IP information in the redis and the in-memory databases
			// for later use
			_ = h.RedisRepo.Save(ctx, g)
			_ = h.InMemoryRepo.Save(ctx, g)
		} else {
			log.Println("cache hit from redis database")
		}
	} else {
		log.Println("cache hit from in-memory database")
	}

	// Marshal the response in json format
	resp, err := json.Marshal(g)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`error: couldn't marshal geo IP information`))
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.Write(resp)
}

// isIpv4 verify the given string is a valid IPv4 address.
// Return true if yes, false otherwise
func isIpv4(host string) bool {
	return net.ParseIP(host) != nil
}
