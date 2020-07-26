package setup

import (
	"context"
	"github.com/jinzhu/gorm"
	"go-workspace/gin/layer/config"
	"go-workspace/gin/layer/serverenv"
	"go-workspace/gin/layer/storage"
)

type Defer func()

type DBConfigProvider interface {
	DB() *storage.Config
}

func Setup(ctx context.Context, config *config.Config) (*serverenv.ServerEnv, Defer, error) {
	var opts []serverenv.Option

	// setup database
	db, err := gorm.Open(config.Database.Endpoint, config.Database.Endpoint)
	if err != nil {
		return nil, func() {}, err
	}
	opts = append(opts, serverenv.WithDatabase(db))

	// setup cache

	return serverenv.NewServerEnv(ctx, opts...), func() {}, nil
}
