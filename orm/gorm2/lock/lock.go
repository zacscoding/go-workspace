package lock

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"
)

var (
	ErrAlreadyLocked = errors.New("lock already in use")
)

type UnlockFn func() error

type Lock struct {
	db *gorm.DB
}

func (l *Lock) Lock(ctx context.Context, lockID string, ttl time.Duration) (UnlockFn, error) {
	var (
		expires sql.NullTime
		opts    = &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  false,
		}
	)

	if err := runInTx(ctx, l.db, opts, func(txDb *gorm.DB) error {
		row := txDb.Raw("CALL AcquireLock(?, ?)", lockID, int32(ttl.Seconds()))

		if err := row.Scan(&expires).Error; err != nil {
			return fmt.Errorf("failed to scan lock.expires: %w", err)
		}

		if !expires.Valid {
			return ErrAlreadyLocked
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return func() error {
		return runInTx(ctx, l.db, opts, func(txDb *gorm.DB) error {
			row := txDb.Raw("CALL ReleaseLock(?, ?)", lockID, expires)
			var released bool
			if err := row.Scan(&released).Error; err != nil {
				return fmt.Errorf("failed to scan lock.released: %w", err)
			}
			return nil
		})
	}, nil
}

func runInTx(ctx context.Context, db *gorm.DB, opts *sql.TxOptions, f func(txDb *gorm.DB) error) error {
	tx := db.WithContext(ctx).Begin(opts)
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
