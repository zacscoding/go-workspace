package resolver

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"testing"
)

/**
Test DBResolver
require master/slave database. see compose/mysql-cluster
*/
func Test1(t *testing.T) {
	db := newDB(t)

	assert.NoError(t, db.Migrator().DropTable(&User{}))
	assert.NoError(t, db.Migrator().AutoMigrate(&User{}))

	fmt.Println("## Try to save a user")
	saveUser := User{
		Name: "user1",
	}
	assert.NoError(t, db.Create(&saveUser).Error)

	fmt.Println("## Try to find a user")
	var find User
	assert.NoError(t, db.First(&find, "name = ?", saveUser.Name).Error)

	fmt.Println("## Try to find a user with exec")
	db.Exec("SELECT * FROM users").Rows()

	fmt.Println("## Try to update a user")
	saveUser.Name = "updated-user1"
	assert.NoError(t, db.Updates(&saveUser).Error)

	fmt.Println("## Try to delete a user")
	assert.NoError(t, db.Delete(&saveUser).Error)
	// ## Try to save a user
	// > Create Callback
	// >> current user: mydb_user@172.27.0.1
	// ## Try to find a user
	// > Query Callback
	// >> current user: mydb_slave_user@172.27.0.1
	// ## Try to find a user with exec
	// > Row Callback
	// >> current user: mydb_slave_user@172.27.0.1
	// ## Try to update a user
	// > Update Callback
	// >> current user: mydb_user@172.27.0.1
	// ## Try to delete a user
	// > Delete Callback
}

func newDB(t *testing.T) *gorm.DB {
	var (
		masterDSN = "mydb_user:mydb_pwd@tcp(127.0.0.1:4406)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
		slaveDSN  = "mydb_slave_user:mydb_slave_pwd@tcp(127.0.0.1:5506)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
	)
	db, err := gorm.Open(mysql.Open(masterDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err)
	db.Use(dbresolver.Register(dbresolver.Config{
		// use `masterDSN` as sources (DB's default connection)
		// `slaveDSN` as replicas
		Replicas: []gorm.Dialector{mysql.Open(slaveDSN)},
	}))

	// TODO : Adds callbacks to check db connection's dsn
	// checked dbresolver.resolve() method
	db.Callback().Create().After("gorm:db_resolver").Register("callback:create", func(db *gorm.DB) {
		fmt.Println("> Create Callback")
		user, err := currentDBUser(db.Statement.ConnPool.(*sql.DB))
		assert.NoError(t, err)
		fmt.Println(">> current user:", user)
	})
	db.Callback().Delete().After("gorm:db_resolver").Register("callback:delete", func(db *gorm.DB) {
		fmt.Println("> Delete Callback")
		user, err := currentDBUser(db.Statement.ConnPool.(*sql.DB))
		assert.NoError(t, err)
		fmt.Println(">> current user:", user)
	})
	db.Callback().Query().After("gorm:db_resolver").Register("callback:query", func(db *gorm.DB) {
		fmt.Println("> Query Callback")
		user, err := currentDBUser(db.Statement.ConnPool.(*sql.DB))
		assert.NoError(t, err)
		fmt.Println(">> current user:", user)
	})
	db.Callback().Row().After("gorm:db_resolver").Register("callback:row", func(db *gorm.DB) {
		fmt.Println("> Row Callback")
		user, err := currentDBUser(db.Statement.ConnPool.(*sql.DB))
		assert.NoError(t, err)
		fmt.Println(">> current user:", user)
	})
	db.Callback().Update().After("gorm:db_resolver").Register("callback:update", func(db *gorm.DB) {
		fmt.Println("> Update Callback")
		user, err := currentDBUser(db.Statement.ConnPool.(*sql.DB))
		assert.NoError(t, err)
		fmt.Println(">> current user:", user)
	})
	return db
}

func currentDBUser(db *sql.DB) (string, error) {
	conn, err := db.Conn(context.Background())
	if err != nil {
		return "", err
	}
	defer conn.Close()

	var user string
	err = conn.QueryRowContext(context.Background(), "SELECT USER()").Scan(&user)
	if err != nil {
		return "", err
	}
	return user, nil
}
