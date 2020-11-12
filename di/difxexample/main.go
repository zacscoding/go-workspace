package main

import (
	"context"
	"go-workspace/di/difx"
	"go.uber.org/fx"
	"log"
	"sync"
)

var (
	runner *Runner
	wg     = sync.WaitGroup{}
)

type Runner struct {
	h *difx.Handler
}

func setupRunner(lc fx.Lifecycle, h *difx.Handler) {
	runner = &Runner{h: h}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			wg.Done()
			return nil
		},
	})
}

func main() {
	runNameDI()
}

func runNameDI() {
	wg.Add(1)
	app := fx.New(
		fx.Provide(
			fx.Annotated{
				Name:   "ro",
				Target: difx.NewReadOnlyDatabase,
			},
			fx.Annotated{
				Name:   "rw",
				Target: difx.NewWriteDatabase,
			},
			difx.NewHandler,
		),
		fx.Invoke(
			setupRunner,
		),
	)
	go func() {
		app.Run()
	}()
	wg.Wait()

	saved := runner.h.SaveMember("member1")
	log.Printf("Success to save a member. id:%d", saved.Id)

	find := runner.h.MemberByID(saved.Id)
	log.Printf("Success to find a member. id:%d, name:%s", find.Id, find.Name)
	// Output
	//2020/11/12 21:15:13 Try to write member from `Write DB`, name:member1
	//2020/11/12 21:15:13 Success to save a member. id:1
	//2020/11/12 21:15:13 Try to read member from `ReadOnly DB`, id:1
	//2020/11/12 21:15:13 Success to find a member. id:1, name:member1
}
