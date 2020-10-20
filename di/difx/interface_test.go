package difx

import (
	"go.uber.org/fx"
	"testing"
)

func TestTemp(t *testing.T) {
	app := fx.New(
		fx.Provide(
			// inmemory db
			NewInmemoryDB,
			// file db
			NewFileDB,
			// db client
			// TODO : named DI?
			NewDatabaseClient,
		),
	)
	app.Run()
}
