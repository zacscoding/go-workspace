package connect

import (
	"database/sql"
	"gorm.io/gorm"
	"testing"
)

const (
	MySqlVersionQuery = "SELECT VERSION()"
)

func TestMysqlNewDatabase(t *testing.T) {
	dsn := "root:password@tcp(127.0.0.1:13306)/my_database?charset=utf8&parseTime=True"
	version := "8.0.17"
	tt := []struct {
		Name            string
		NewDatabaseFunc func(dsn string) (*gorm.DB, error)
	}{
		{
			Name: "NewDatabase",
			NewDatabaseFunc: func(dsn string) (*gorm.DB, error) {
				return NewMysqlDatabase(dsn)
			},
		}, {
			Name: "NewDatabaseWithConfig",
			NewDatabaseFunc: func(dsn string) (*gorm.DB, error) {
				return NewMysqlDatabaseWithConfig(dsn)
			},
		}, {
			Name: "NewDatabaseWithConfig",
			NewDatabaseFunc: func(dsn string) (*gorm.DB, error) {
				db, _ := sql.Open("mysql", dsn)
				return NewDatabaseWithConnection(db)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			db, err := tc.NewDatabaseFunc(dsn)
			if err != nil {
				t.Errorf("wanted nil err, got: %v", err)
				t.Fail()
			}

			var v string
			rdb, _ := db.DB()
			err = rdb.QueryRow(MySqlVersionQuery).Scan(&v)
			if err != nil {
				t.Errorf("version query got err: %v", err)
				t.Fail()
			}
			if version != v {
				t.Errorf("version expected:%s, got:%s", version, v)
				t.Fail()
			}
		})
	}
}
