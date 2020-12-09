package tx

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
)

type AccountModel struct {
	Email    string `gorm:"column:email;type:VARCHAR(80);UNIQUE_INDEX:idx_unique_email_name;NOT NULL"`
	Username string `gorm:"column:username;type:VARCHAR(80);UNIQUE_INDEX:idx_unique_email_name;NOT NULL"`
	Age      int    `gorm:"column:age;type:int;"`
}

func InTxWithContext(ctx context.Context, db *gorm.DB, opts *sql.TxOptions, f func(txDB *gorm.DB) error) error {
	tx := db.BeginTx(ctx, opts)
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
