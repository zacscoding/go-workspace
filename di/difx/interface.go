package difx

import "fmt"

// Database

type Database interface {
	Save() error
}

type inmemoryDB struct {
}

func (idb *inmemoryDB) Save() error {
	return nil
}

type fileDB struct {
}

func (idb *fileDB) Save() error {
	return nil
}

func NewInmemoryDB() Database {
	return &inmemoryDB{}
}

func NewFileDB() Database {
	return &fileDB{}
}

// Database client
type DatabaseClient struct {
	DB Database
}

func NewDatabaseClient(db Database) *DatabaseClient {
	switch db.(type) {
	case *inmemoryDB:
		fmt.Println("NewDatabaseClient with inmemoryDB")
	case *fileDB:
		fmt.Println("NewDatabaseClient with fileDB")
	default:
		fmt.Println("NewDatabaseClient with unknown db")
	}
	return &DatabaseClient{DB: db}
}
