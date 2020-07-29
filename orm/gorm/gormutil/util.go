package gormutil

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

func InTx(db *gorm.DB, f func(rdb *gorm.DB) error) error {
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("start tx: %v", tx.Error)
	}

	if err := f(tx); err != nil {
		if err1 := tx.Rollback().Error; err1 != nil {
			return fmt.Errorf("rollback tx: %v (original error: %v)", err1, err)
		}
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit tx: %v", err)
	}
	return nil
}
