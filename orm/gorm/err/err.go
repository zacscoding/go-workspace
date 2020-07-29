package err

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open("mysql", "root:password@tcp(127.0.0.1:13306)/my_database?charset=utf8&parseTime=True")
	if err != nil {
		panic(err)
	}
	db.LogMode(true)
	db.DropTableIfExists(&Account{})
	db.AutoMigrate(&Account{})
}

type Account struct {
	gorm.Model
	Name1 string  `gorm:"unique;not null"`
	Name2 *string `gorm:"not null"`
}
