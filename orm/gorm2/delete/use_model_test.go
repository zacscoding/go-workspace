package delete

import (
	mysql2 "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"testing"
	"time"
)

type UserWithModel struct {
	gorm.Model
	Name string `json:"name" gorm:"unique"`
}

func TestUserWithModel(t *testing.T) {
	db := NewDB(t)

	//----------------------------------------------
	// Find function
	// > Query: SELECT * FROM `user_with_models` WHERE `user_with_models`.`deleted_at` IS NULL
	//----------------------------------------------
	var findUsers []*UserWithModel
	err := db.Find(&findUsers).Error
	assert.NoError(t, err)

	//----------------------------------------------
	// Save function
	// > Query: INSERT INTO `user_with_models` (`created_at`,`updated_at`,`deleted_at`,`name`) VALUES ('2021-08-22 01:37:10.026','2021-08-22 01:37:10.026',NULL,'user1')
	//----------------------------------------------
	user1 := UserWithModel{Name: "user1"}
	err = db.Save(&user1).Error
	assert.NoError(t, err)

	//----------------------------------------------
	// Update function
	// > Query: UPDATE `user_with_models` SET `created_at`='2021-08-22 02:00:05.063',`updated_at`='2021-08-22 02:00:05.068',`deleted_at`=NULL,`name`='updated-user1' WHERE `id` = 1
	//----------------------------------------------
	user1.Name = "updated-user1"
	err = db.Save(&user1).Error
	assert.NoError(t, err)

	//----------------------------------------------
	// First function
	// > Query: SELECT * FROM `user_with_models` WHERE `user_with_models`.`id` = 1 AND `user_with_models`.`deleted_at` IS NULL ORDER BY `user_with_models`.`id` LIMIT 1
	//----------------------------------------------
	var findUser UserWithModel
	err = db.First(&findUser, user1.ID).Error

	//----------------------------------------------
	// Delete function
	// > Query: SELECT * FROM `user_with_models` WHERE `user_with_models`.`id` = 1 AND `user_with_models`.`deleted_at` IS NULL ORDER BY `user_with_models`.`id` LIMIT 1
	//----------------------------------------------
	err = db.Delete(&findUser).Error
	assert.NoError(t, err)

	//----------------------------------------------
	// Save function
	// > Query: INSERT INTO `user_with_models` (`created_at`,`updated_at`,`deleted_at`,`name`) VALUES ('2021-08-22 01:42:03.403','2021-08-22 01:42:03.403',NULL,'user1')
	//----------------------------------------------
	user2 := UserWithModel{Name: user1.Name}
	err = db.Save(&user2).Error
	assert.Error(t, err)
	merr, ok := err.(*mysql2.MySQLError)
	assert.True(t, ok)
	assert.EqualValues(t, 1062, merr.Number)
	assert.Contains(t, merr.Message, "Duplicate")
}

func NewDB(t *testing.T) *gorm.DB {
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
		t.Fatalf("open db:%v", err)
	}
	models := []interface{}{
		&UserWithModel{},
		&UserWithSoftDelete{},
		&UserWithDeleteFlag{},
		&UserWithPointer{},
	}

	if err := db.Migrator().DropTable(models...); err != nil {
		t.Fatalf("drop tables:%v", err)
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate tables:%v", err)
	}
	return db
}
