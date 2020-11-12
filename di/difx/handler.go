package difx

import (
	"go.uber.org/fx"
	"log"
)

type HandlerParams struct {
	fx.In
	WriteDatabase Database `name:"rw"`
	ReadDatabase  Database `name:"ro" optional:"true"`
}

type Handler struct {
	readDB  Database
	writeDB Database
}

func (h *Handler) MemberByID(id int64) *Member {
	log.Printf("Try to read member from `%s`, id:%d", h.readDB.GetDatabaseName(), id)
	return h.readDB.GetMemberById(id)
}

func (h *Handler) SaveMember(name string) *Member {
	log.Printf("Try to write member from `%s`, name:%s", h.writeDB.GetDatabaseName(), name)
	return h.writeDB.SaveMember(&Member{Name: name})
}

func NewHandler(p HandlerParams) *Handler {
	log.Println("##[ENV] NewHandler")
	h := Handler{
		writeDB: p.WriteDatabase,
	}

	if p.ReadDatabase == nil {
		log.Println("Configure WriteDatabase to ReadDatabae because nil")
		h.readDB = h.writeDB
	} else {
		log.Printf("Configure ReadDatabase: %s", p.ReadDatabase.GetDatabaseName())
		h.readDB = p.ReadDatabase
	}
	return &h
}
