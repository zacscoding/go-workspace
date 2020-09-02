package connect

import (
	"database/sql"
	gMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewMysqlDatabase creates a new mysql gorm.DB with given datasource name
func NewMysqlDatabase(dsn string) (*gorm.DB, error) {
	return gorm.Open(gMysql.Open(dsn), &gorm.Config{})
}

// NewMysqlDatabaseWithConfig creates a new mysql gorm.DB with given datasource name and mysql config
func NewMysqlDatabaseWithConfig(dsn string) (*gorm.DB, error) {
	return gorm.Open(gMysql.New(gMysql.Config{
		DSN:                       dsn,   // data source name
		DefaultStringSize:         256,   // default size for string fields
		DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
	}), &gorm.Config{})
}

// NewDatabaseWithConnection creates a new mysql gorm.DB with given existing database connections
func NewDatabaseWithConnection(db *sql.DB) (*gorm.DB, error) {
	return gorm.Open(gMysql.New(gMysql.Config{
		Conn: db,
	}), &gorm.Config{})
}
