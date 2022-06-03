package controllers

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
