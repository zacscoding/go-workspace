package difx

import (
	"log"
	"sync"
	"sync/atomic"
)

var (
	idGenerator int64 = 0
	members           = make(map[int64]*Member)
	mutex       sync.RWMutex
)

type Member struct {
	Id   int64
	Name string
}

type Database interface {
	SaveMember(member *Member) *Member

	GetMemberById(id int64) *Member

	GetDatabaseName() string
}

type Reader interface {
	GetMemberById(id int64) *Member
}

type Writer interface {
	SaveMember(member *Member) *Member
}

type inmemoryDB struct {
	readDB  Reader
	writeDB Writer
	name    string
}

func (i *inmemoryDB) SaveMember(member *Member) *Member {
	mutex.Lock()
	defer mutex.Unlock()
	m := Member{
		Id:   atomic.AddInt64(&idGenerator, 1),
		Name: member.Name,
	}
	members[m.Id] = &m
	member.Id = m.Id
	return member
}

func (i *inmemoryDB) GetDatabaseName() string {
	return i.name
}

func (i *inmemoryDB) GetMemberById(id int64) *Member {
	mutex.RLock()
	defer mutex.RUnlock()
	m, ok := members[id]
	if !ok {
		return nil
	}
	return &Member{
		Id:   m.Id,
		Name: m.Name,
	}
}

func NewReadOnlyDatabase() Database {
	log.Println("##[ENV] NewReadOnlyDatabase")
	return &inmemoryDB{
		name: "ReadOnly DB",
	}
}

func NewWriteDatabase() Database {
	log.Println("##[ENV] NewWriteDatabase")
	return &inmemoryDB{
		name: "Write DB",
	}
}
