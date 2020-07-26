package config

import (
	"go-workspace/gin/layer/storage"
)

type Config struct {
	Database *storage.Config
}
