package models

// ErrorResponse defines the standard error response structure.
// @Description Error response structure with a message.
type ErrorResponse struct {
	Message string `json:"message,omitempty"`
}
