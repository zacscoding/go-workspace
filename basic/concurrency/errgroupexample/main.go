package main

import (
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"log"
	"time"
)

func main() {
	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		time.Sleep(time.Second * 1)
		log.Println("terminated fn1")
		//return errors.New("fn1")
		return nil
	})

	g.Go(func() error {
		time.Sleep(time.Second * 2)
		log.Println("terminated fn2")
		return errors.New("fn2")
	})

	g.Go(func() error {
		time.Sleep(time.Second * 3)
		log.Println("terminated fn3")
		return errors.New("fn3")
	})

	err := g.Wait()
	log.Println("Err: ", err)
}
