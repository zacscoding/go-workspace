package main

import (
	"context"
	"errors"
	"github.com/hibiken/asynq"
	"log"
	"time"
)

const (
	redisAddr = "localhost:6379"
)

const (
	TypeEmailDelivery = "email:delivery"
)

func main() {
	err := produce()
	log.Println("Produce task:", err)

	r := asynq.RedisClientOpt{Addr: redisAddr}
	srv := asynq.NewServer(r, asynq.Config{
		// Specify how many concurrent workers to use
		Concurrency: 10,
		// Optionally specify multiple queues with different priority.
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
		// See the godoc for other configuration options
	})
	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeEmailDelivery, func(ctx context.Context, task *asynq.Task) error {
		log.Println("Do Task..", task)
		return errors.New("force error")
	})

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}

func produce() error {
	r := asynq.RedisClientOpt{Addr: redisAddr}
	c := asynq.NewClient(r)
	defer c.Close()

	payload := map[string]interface{}{"user_id": "user1", "template_id": "template1"}
	t := asynq.NewTask(TypeEmailDelivery, payload)
	_, err := c.Enqueue(t,
		asynq.ProcessAt(time.Now().Add(time.Minute)),
		asynq.MaxRetry(5),
	)
	return err
}
