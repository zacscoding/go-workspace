package health

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

type RedisIndicator struct {
	rdb redis.UniversalClient
}

func (r *RedisIndicator) Health(ctx context.Context) Health {
	h := NewHealth()

	err := r.rdb.Ping(ctx).Err()
	if err != nil {
		return *h.WithDown().WithDetail("err", err.Error())
	}

	// TODO : check
	switch r.rdb.(type) {
	case *redis.ClusterClient:
		fmt.Println("Redis cluster client")

		client := r.rdb.(*redis.ClusterClient)
		clusterInfo, _ := client.ClusterInfo(ctx).Result()

		fmt.Println("Cluster info")
		fmt.Println(clusterInfo)
	case *redis.Client:
		fmt.Println("Redis client")

		client := r.rdb.(*redis.Client)
		version, _ := client.Info(ctx, "server.redis_version").Result()

		fmt.Println("Info")
		fmt.Println(version)
	}
	return *h.WithUp()
}

func NewRedisIndicator(rdb redis.UniversalClient) Indicator {
	return &RedisIndicator{rdb: rdb}
}
