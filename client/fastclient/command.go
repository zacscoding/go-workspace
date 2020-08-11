package fastclient

import (
	"encoding/json"
	"github.com/sony/gobreaker"
	"log"
	"time"
)

type Command interface {
	Execute(req func() error) error
}

type NoopCommand struct {
}

func (n *NoopCommand) Execute(req func() error) error {
	return req()
}

type CircuitBreakerCommand struct {
	cb *gobreaker.CircuitBreaker
}

func (c *CircuitBreakerCommand) Execute(req func() error) error {
	_, err := c.cb.Execute(func() (_ interface{}, err error) {
		return nil, req()
	})
	return err
}

func NewCircuitBreakerCommand(name string) *CircuitBreakerCommand {
	st := gobreaker.Settings{
		Name:        name,
		MaxRequests: 0,
		Interval:    0,
		Timeout:     2 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			b, _ := json.Marshal(counts)
			log.Printf("ReadyToTrip: %s", string(b))
			rate := float32(counts.ConsecutiveFailures) / float32(counts.Requests)
			if rate >= 0.5 {
				return true
			}
			return false
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("Circuit[%s] %v -> %v", name, from, to)
		},
	}
	return &CircuitBreakerCommand{cb: gobreaker.NewCircuitBreaker(st)}
}
