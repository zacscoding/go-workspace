package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"go-workspace/health"
	"time"
)

func main() {
	indicators := make(map[string]health.Indicator)

	// 1) mysql
	//mysqlDb, _ := sql.Open("mysql", "root:password@tcp(127.0.0.1:13306)/my_database?charset=utf8&parseTime=True")
	//defer mysqlDb.Close()
	//indicators["mysql"] = health.NewDbHealthChecker(mysqlDb, "MySQL", "SELECT 1", "SELECT VERSION()")

	// 2) redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	indicators["redis-cluster"] = health.NewRedisIndicator(rdb)

	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			for name, indicator := range indicators {
				fmt.Println("Check", name)
				h := indicator.Health(ctx)
				b, _ := json.Marshal(h)
				fmt.Println(string(b))
			}
			cancel()
		}
	}
}
