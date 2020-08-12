package save

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSaveIfNotExist(t *testing.T) {
	name := "user1"

	_, _ = SaveIfNotExist(name)
	_, _ = SaveIfNotExist(name)
}

func TestSaveOrUpdate(t *testing.T) {
	acc := Account{
		Name: "acc1",
	}

	err := SaveOrUpdate(&acc)
	if err != nil {
		panic(err)
	}

	acc2 := Account{
		Model: gorm.Model{
			ID: acc.ID,
		},
		Name: "newacc1",
	}

	err = SaveOrUpdate(&acc2)
	if err != nil {
		panic(err)
	}
}

func TestSave(t *testing.T) {
	acc := Account{
		Name: "account1",
		Age:  10,
	}

	// 1) save
	chain := db.Create(&acc)
	assert.NoError(t, chain.Error)
	assert.Equal(t, int64(1), chain.RowsAffected)
	b, _ := json.Marshal(acc)
	fmt.Println("acc ::", string(b))

	// 2) update
	updatedId := acc.ID
	updatedName := "new-account"
	updatedAcc := Account{
		Model: gorm.Model{
			ID:        updatedId + 1,
			UpdatedAt: time.Now(),
		},
		Name: updatedName,
	}

	// chain = db.Model(&Account{}).Where("id = ?", updatedAcc.ID).UpdateColumns(&updatedAcc)
	//db.Model(&Account{}).Where()

	chain = db.Save(&updatedAcc)
	assert.NoError(t, chain.Error)
	assert.Equal(t, int64(1), chain.RowsAffected)
}
