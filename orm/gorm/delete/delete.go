package delete

import (
	"errors"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"go-workspace/orm/gorm/gormutil"
	"time"
)

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open("mysql", "root:password@tcp(127.0.0.1:13306)/my_database?charset=utf8&parseTime=True")
	if err != nil {
		panic(err)
	}
	db.LogMode(true)
	db.DropTableIfExists(&Account{}, &Article{})
	db.AutoMigrate(&Account{}, &Article{})
}

type Account struct {
	gorm.Model
	Name     string `gorm:"unique_index; not null"`
	Articles []Article
}

type Article struct {
	gorm.Model
	Title     string
	Account   Account
	AccountID uint
}

func DeleteAccount(accountId uint, fail bool) error {
	return gormutil.InTx(db, func(rdb *gorm.DB) error {
		// delete account
		query := "UPDATE accounts SET deleted_at = ? WHERE id = ?"
		err := db.Exec(query, time.Now(), accountId).Error
		if err != nil {
			return err
		}
		// force error
		if fail {
			return errors.New("force exception")
		}

		// delete account's articles
		query = "UPDATE articles SET deleted_at = ? WHERE account_id = ?"
		return db.Exec(query, time.Now(), accountId).Error
	})
}
