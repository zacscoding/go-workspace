package hook

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"testing"
	"time"
)

func TestOne(t *testing.T) {
	db := newDB(t)
	assert.NoError(t, db.Migrator().AutoMigrate(&HookUser{}))
	for i := 0; i < 10; i++ {
		u := HookUser{
			Name:        fmt.Sprintf("user-%d", i+1),
			beforeSleep: time.Second,
			afterSleep:  time.Second,
		}
		assert.NoError(t, db.Create(&u).Error)
	}
}

func TestSaveBulk(t *testing.T) {
	db := newDB(t)
	assert.NoError(t, db.Migrator().AutoMigrate(&HookUser{}))

	var users []*HookUser
	for i := 0; i < 10; i++ {
		u := &HookUser{
			Name:        fmt.Sprintf("user-%d", i+1),
			beforeSleep: time.Second,
			afterSleep:  time.Second,
		}
		if i == 7 {
			u.afterSleep = time.Minute * 6
		}
		users = append(users, u)
	}
	db.Create(users)
}

func newDB(t *testing.T) *gorm.DB {
	dsn := "root:password@tcp(127.0.0.1:13306)/my_database?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second,   // Slow SQL threshold
				LogLevel:      logger.Silent, // Log level
				Colorful:      true,
			},
		),
	})
	if err != nil || db == nil {
		t.Fatalf("open db:%v", err)
	}
	return db
}

func Test2(t *testing.T) {
	maxPlaceholders := 65535
	placeholdersPerTxItem := 13
	log.Println(maxPlaceholders / placeholdersPerTxItem)
}
