package selectforupdate

import (
	"github.com/jinzhu/gorm"
	"math/rand"
	"time"
)

type Data struct {
	gorm.Model
	SyncColumn1 bool `gorm:"column:sync_column1;type:TINYINT(1);;DEFAULT:0"`
	SyncColumn2 bool `gorm:"column:sync_column2;type:TINYINT(1);;DEFAULT:0"`
	Completed   bool `gorm:"column:completed;type:TINYINT(1);;DEFAULT:0"`
}

type Repository struct {
	db *gorm.DB
}

func sleepRandMills(n int) {
	time.Sleep(time.Duration(rand.Intn(n)) * time.Millisecond)
}

func (r *Repository) UpdateSyncColumn1(id int64) (bool, error) {
	sleepRandMills(100)
	executed := r.db.Model(new(Data)).Where("id = ?", id).UpdateColumns(map[string]interface{}{
		"sync_column1": true,
	})
	if executed.Error != nil {
		return false, executed.Error
	}
	if err := r.UpdateCompleted(id); err != nil {
		return false, err
	}
	return executed.RowsAffected == 1, nil
}

func (r *Repository) UpdateSyncColumn2(id int64) (bool, error) {
	sleepRandMills(100)
	executed := r.db.Model(new(Data)).Where("id = ?", id).UpdateColumns(map[string]interface{}{
		"sync_column2": true,
	})
	if executed.Error != nil {
		return false, executed.Error
	}
	if err := r.UpdateCompleted(id); err != nil {
		return false, err
	}
	return executed.RowsAffected == 1, nil
}

func (r *Repository) UpdateCompleted(id int64) error {
	var data Data
	if err := r.db.Model(new(Data)).First(&data, "id = ?", id).Error; err != nil {
		return err
	}
	if !data.SyncColumn1 || !data.SyncColumn2 {
		//fmt.Printf("[#%d] not synchronized yet. col1: %v, col2:%v\n", data.ID, data.SyncColumn1, data.SyncColumn2)
		return nil
	}
	return r.db.Model(new(Data)).Where("id = ?", id).UpdateColumns(map[string]interface{}{
		"completed": true,
	}).Error
}
