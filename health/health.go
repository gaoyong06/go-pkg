package health

import "time"

const (
	// StatusUp indicates the service is running normally.
	StatusUp = "UP"
)

// Response defines a unified payload for service health checks.
type Response struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp int64  `json:"timestamp"`
}

// NewResponse builds a standard health response for the given service.
func NewResponse(service string) *Response {
	return &Response{
		Status:    StatusUp,
		Service:   service,
		Timestamp: time.Now().Unix(),
	}
}
