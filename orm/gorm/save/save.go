package save

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
	Name string `gorm:"unique_index; not null"`
}

func SaveOrUpdate(acc *Account) error {
	return db.Save(acc).Error
}

func SaveIfNotExist(name string) (*Account, error) {
	var acc Account
	err := db.Where("name = ?", name).First(&acc).Error
	// exist
	if err == nil {
		return &acc, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	acc.Name = name
	if err = db.Create(&acc).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}
