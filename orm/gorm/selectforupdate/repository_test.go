package selectforupdate

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func Test1(t *testing.T) {
	for repeat := 0; repeat < 50; repeat++ {
		db := newDatabase(t, 1, 50)
		repo := &Repository{db: db}
		wg := sync.WaitGroup{}

		for i := 1; i <= 50; i++ {
			i := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				repo.UpdateSyncColumn1(int64(i))
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				repo.UpdateSyncColumn2(int64(i))
			}()
		}
		wg.Wait()
		var data []*Data
		assert.NoError(t, db.Model(new(Data)).Find(&data).Error)
		for _, d := range data {
			if !d.SyncColumn1 || !d.SyncColumn2 || !d.Completed {
				fmt.Printf("[Repeat-%d-%d] not synchronized yet. col1: %v, col2:%v, completed:%v\n",
					repeat, d.ID, d.SyncColumn1, d.SyncColumn2, d.Completed)
			}
		}
		db.Close()
	}
}

func newDatabase(t *testing.T, start, last uint) *gorm.DB {
	db, err := gorm.Open("mysql", "root:password@tcp(127.0.0.1:13306)/my_database?charset=utf8&parseTime=True")
	assert.NoError(t, err)
	assert.NoError(t, db.DropTableIfExists(&Data{}).Error)
	assert.NoError(t, db.AutoMigrate(&Data{}).Error)
	for i := start; i <= last; i++ {
		data := Data{
			Model: gorm.Model{
				ID: i,
			},
			SyncColumn1: false,
			SyncColumn2: false,
			Completed:   false,
		}
		assert.NoError(t, db.Save(&data).Error)
	}
	return db
}
