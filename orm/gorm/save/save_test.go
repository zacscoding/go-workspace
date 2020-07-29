package save

import (
	"github.com/jinzhu/gorm"
	"testing"
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
