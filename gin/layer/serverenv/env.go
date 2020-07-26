package serverenv

import (
	"context"
	"github.com/jinzhu/gorm"
)

type ServerEnv struct {
	database *gorm.DB
}

type Option func(*ServerEnv) *ServerEnv

// NewServerEnv create a new ServerEnv with given options
func NewServerEnv(ctx context.Context, opts ...Option) *ServerEnv {
	env := &ServerEnv{}

	for _, f := range opts {
		env = f(env)
	}

	return env
}

func WithDatabase(db *gorm.DB) Option {
	return func(s *ServerEnv) *ServerEnv {
		s.database = db
		return s
	}
}
