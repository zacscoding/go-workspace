package main

import (
	"context"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"go-workspace/poc/cachedb/user/database"
	"go-workspace/poc/cachedb/user/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func main() {
	db := NewUserDB(true)

	if _, ok := db.(database.UserCacheEvictor); ok {
		log.Printf("Possible to cast UserCacheEvictor.")
	}

	u1 := model.User{
		Email:    "user1@gmail.com",
		Name:     "user1",
		Password: "user1pass",
		Bio:      "user1bio",
		Image:    "user1image",
	}

	checkFindAll(db)

	if err := db.Save(context.Background(), &u1); err != nil {
		log.Fatal(err)
	}

	if find, err := db.FindByID(context.Background(), u1.ID); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Find user: %s", find.String())
	}
	checkFindAll(db)

	log.Println("Sleep 10 seconds..")
	if err := db.Update(context.Background(), &model.User{
		ID:       u1.ID,
		Email:    "user1@gmail.com",
		Name:     "UpdateUser1",
		Password: "user1pass",
		Bio:      "user1bio",
		Image:    "user1image",
	}); err != nil {
		log.Fatal(err)
	}
	checkFindAll(db)
}

func checkFindAll(db database.UserDB) {
	users, err := db.FindAll(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("## Find users: %d", len(users))
	for _, user := range users {
		log.Printf("> User: %s", user)
	}
}

func NewUserDB(cacheEnabled bool) database.UserDB {
	// (1) initialize gorm
	dsn := "root:password@tcp(127.0.0.1:13306)/my_database?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second, // Slow SQL threshold
				LogLevel:      logger.Info, // Log level
				Colorful:      true,
			},
		),
	})
	if err != nil || db == nil {
		log.Fatal(err)
	}
	db.Migrator().DropTable(&model.User{})
	if err := db.Migrator().AutoMigrate(&model.User{}); err != nil {
		log.Fatal(err)
	}

	userDB := database.NewUserDB(db)
	if !cacheEnabled {
		return userDB
	}

	// (2) setup redis
	cli := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			"localhost:7000", "localhost:7001", "localhost:7002", "localhost:7003", "localhost:7004", "localhost:7005",
		},
		ReadTimeout:   3 * time.Second,
		WriteTimeout:  5 * time.Second,
		DialTimeout:   5 * time.Second,
		ReadOnly:      true,
		RouteRandomly: true,
	})
	rcache := cache.New(&cache.Options{
		Redis: cli,
	})

	userCacheDB := database.NewUserCache(rcache, userDB)
	return userCacheDB
}
