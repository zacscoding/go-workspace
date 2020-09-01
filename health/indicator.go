package health

import (
	"context"
	"encoding/json"
)

type status string

const (
	up      status = "UP"
	down           = "DOWN"
	unknown        = "UNKNOWN"
)

type Health struct {
	status  status
	details map[string]interface{}
}

func (h Health) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{}
	data["status"] = h.status
	data["details"] = h.details
	return json.Marshal(data)
}

func (h *Health) WithUp() *Health {
	h.status = up
	return h
}

func (h *Health) WithDown() *Health {
	h.status = down
	return h
}

func (h *Health) WithDetail(key string, value interface{}) *Health {
	h.details[key] = value
	return h
}

type Indicator interface {
	Health(ctx context.Context) Health
}

// NewHealth creates a new health state
func NewHealth() Health {
	return Health{
		status:  unknown,
		details: make(map[string]interface{}),
	}
}
