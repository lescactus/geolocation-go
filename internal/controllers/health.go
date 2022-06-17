package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
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
	Status string `json:"status"`
	Checks struct {
		InMemoryRepoStatus string `json:"in_memory_database_status"`
		RedisRepoStatus    string `json:"redis_database_status"`
		RemoteIPAPIStatus  string `json:"remote_api_status"`
	} `json:"checks"`
}

// Healthz provide a handler used for health checks.
// TODO: Better implementation
func (h *BaseHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup

	var status string
	var inMemoryRepoStatus string
	var redisRepoStatus string
	var remoteIPAPIStatus string

	var errInMemoryRepo error
	var errRedisRepo error
	var errRemoteIPAPI error

	chErrInMemoryRepoStatus := make(chan error, 1)
	chErrRedisRepoStatus := make(chan error, 1)
	chErrRemoteIPAPIStatus := make(chan error, 1)

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	wg.Add(3)

	go h.InMemoryRepo.Status(ctx, &wg, chErrInMemoryRepoStatus)
	go h.RedisRepo.Status(ctx, &wg, chErrRedisRepoStatus)
	go h.RemoteIPAPI.Status(ctx, &wg, chErrRemoteIPAPIStatus)

	wg.Wait()

	errInMemoryRepo = <-chErrInMemoryRepoStatus
	if errInMemoryRepo != nil {
		inMemoryRepoStatus = HealthzKO
	} else {
		inMemoryRepoStatus = HealthzOK
	}

	errRedisRepo = <-chErrRedisRepoStatus
	if errRedisRepo != nil {
		redisRepoStatus = HealthzKO
	} else {
		redisRepoStatus = HealthzOK
	}

	errRemoteIPAPI = <-chErrRemoteIPAPIStatus
	if errRemoteIPAPI != nil {
		remoteIPAPIStatus = HealthzKO
	} else {
		remoteIPAPIStatus = HealthzOK
	}

	if (inMemoryRepoStatus != HealthzOK) || (redisRepoStatus != HealthzOK) || (remoteIPAPIStatus != HealthzOK) {
		status = HealthzKO
	} else {
		status = HealthzOK
	}

	z := HealthzResponse{
		Status: status,
		Checks: struct {
			InMemoryRepoStatus string `json:"in_memory_database_status"`
			RedisRepoStatus    string `json:"redis_database_status"`
			RemoteIPAPIStatus  string `json:"remote_api_status"`
		}{
			InMemoryRepoStatus: inMemoryRepoStatus,
			RedisRepoStatus:    redisRepoStatus,
			RemoteIPAPIStatus:  remoteIPAPIStatus,
		},
	}

	resp, err := json.Marshal(&z)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.Write(resp)
}
