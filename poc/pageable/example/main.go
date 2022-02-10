package main

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"go-workspace/poc/pageable/app"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"math/rand"
	"os"
	"time"
)

/**
curl -XGET http://localhost:8899/users?lastName=Hirthe
curl -XGET http://localhost:8899/users?lastName=Hirthe&page=1&size=3&sort=name,asc
curl -XGET http://localhost:8899/users?lastName=Hirthe&page=2&size=3&sort=email,desc

*/
func main() {
	var err error
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
	if err != nil {
		panic(err)
	}
	setupFixtures(db)

	handler := app.NewHandler(app.NewUserStore(db))

	e := gin.New()
	e.GET("/users", handler.HandleGetUsers)

	if err := e.Run(":8899"); err != nil {
		log.Fatal(err)
	}
}

func setupFixtures(db *gorm.DB) {
	db.Migrator().DropTable(&app.User{})
	db.Migrator().AutoMigrate(&app.User{})

	lastNames := []string{
		"Hirthe",
		"Schaefer",
		"Wyman",
		"Kassulke",
		"Bartell",
		"Kilback",
		"Harber",
		"Hilpert",
		"Pagac",
		"Fahey",
	}

	var users []*app.User
	for i := 0; i < 100; i++ {
		users = append(users, &app.User{
			FirstName: gofakeit.FirstName(),
			LastName:  lastNames[rand.Intn(100)%len(lastNames)],
			Email:     gofakeit.Email(),
		})
	}
	if err := db.Create(&users).Error; err != nil {
		log.Fatal(err)
	}
}
