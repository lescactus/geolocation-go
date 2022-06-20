package controllers

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents the json response
// for http errors
type ErrorResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

func NewErrorResponse(msg string) *ErrorResponse {
	return &ErrorResponse{
		Status: "error",
		Msg:    msg,
	}
}

// NotFoundHandler is the custom http handler for 404 responses.
// It respond to the client with a 404 status code and a custom
// json error message.
func (h *BaseHandler) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	e := NewErrorResponse("404 page not found")
	resp, _ := json.Marshal(e)

	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusNotFound)
	w.Write(resp)
}

// NotFoundHandler is the custom http handler for 405 responses.
// It respond to the client with a 405 status code and a custom
// json error message.
func (h *BaseHandler) MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	e := NewErrorResponse("405 method not allowed")
	resp, _ := json.Marshal(e)

	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write(resp)
}

// HandleOPTIONS is the custom http handler for OPTIONS requests.
// It respond to the client with a 200 status code and set the
// "Allow" http header.
func (h *BaseHandler) OptionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "GET, OPTIONS")
	w.WriteHeader(http.StatusOK)
}
