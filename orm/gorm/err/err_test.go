package err

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"reflect"
	"testing"
)

func TestDuplicateKeyError(t *testing.T) {
	name := "name1"
	acc := Account{
		Name1: name,
		Name2: &name,
	}
	if err := db.Create(&acc).Error; err != nil {
		panic(err)
	}

	err := db.Create(&acc).Error
	displayDbError(err)
	// Output
	// *mysql.MySQLError >> number: 1062 , message: Duplicate entry '1' for key 'PRIMARY'

	newAcc := Account{
		Name1: acc.Name1,
		Name2: &name,
	}
	err = db.Create(&newAcc).Error

	displayDbError(err)
	// Output
	// *mysql.MySQLError >> number: 1062 , message: Duplicate entry 'name1' for key 'name1'
}

func TestNotNullColumn(t *testing.T) {
	name := "name1"
	acc1 := Account{
		Name2: &name,
	}
	// no error about Name1 having not null tag, because default string is ""
	if err := db.Create(&acc1).Error; err != nil {
		panic(err)
	}

	// occur error because Name2's field is string pointer
	acc2 := Account{
		Name1: name,
	}
	err := db.Create(&acc2).Error
	displayDbError(err)
	// Output
	// *mysql.MySQLError >> number: 1048 , message: Column 'name2' cannot be null
}

func displayDbError(err error) {
	if err == nil {
		fmt.Println("Empty error")
		return
	}

	switch err.(type) {
	case *mysql.MySQLError:
		e := err.(*mysql.MySQLError)
		fmt.Println("*mysql.MySQLError >> number:", e.Number, ", message:", e.Message)
	default:
		fmt.Println("unknown error type:", reflect.TypeOf(err))
	}
}
