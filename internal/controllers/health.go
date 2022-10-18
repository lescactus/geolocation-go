package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const (
	// HealthzOK is the "ok" message for the health endpoint
	HealthzOK = "pass"

	// HealthzKO is the "ko" message for the health endpoint
	HealthzKO = "fail"
)

// HealthzResponse represents the json response of a health endpoint.
// It provides the status of the app dependencies.
type HealthzResponse struct {
	Status string             `json:"global_status"`
	Checks []CacheHealthCheck `json:"checks"`
}

// CacheHealthCheck represents the health status of a cache.
type CacheHealthCheck struct {
	CacheName      string `json:"cache"`
	CacheStatus    string `json:"status"`
	CacheStatusMsg string `json:"msg"`
}

// newChecks convert a map of [string]error to a []HealthCheck
func newChecks(m map[string]error) []CacheHealthCheck {
	if len(m) <= 0 {
		return []CacheHealthCheck{}
	}

	var c = make([]CacheHealthCheck, len(m))

	i := 0
	for name, err := range m {
		var status string
		var msg string
		if err == nil {
			status = HealthzOK
			msg = "alive"
		} else {
			status = HealthzKO
			msg = err.Error()
		}

		c[i] = CacheHealthCheck{
			CacheName:      name,
			CacheStatus:    status,
			CacheStatusMsg: msg,
		}

		i++
	}

	return c
}

// Healthz provide a handler used for health checks.
// It will verify the status of all the cache registries in the cache chain
// and will always answer with a 200 http status code with a HealthzResponse.
func (h *BaseHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	errors := h.CacheChain.Statuses(ctx)

	var status string
	if hasErrors(errors) {
		status = HealthzKO
	} else {
		status = HealthzOK
	}

	health := HealthzResponse{
		Status: status,
		Checks: newChecks(errors),
	}

	resp, err := json.Marshal(&health)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.Write(resp)
}

// hasErrors will return true if the map contains any error(s)
func hasErrors(e map[string]error) bool {
	for _, err := range e {
		if err != nil {
			return true
		}
	}
	return false
}
