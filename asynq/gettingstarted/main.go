package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

// A list of task types.
const (
	TypeEmailDelivery = "email:deliver"
	TypeImageResize   = "image:resize"
)

//----------------------------------------------
// Write a function NewXXXTask to create a task.
// A task consists of a type and a payload.
//----------------------------------------------

func NewEmailDeliveryTask(userID int, tmplID string) *asynq.Task {
	payload := map[string]interface{}{"user_id": userID, "template_id": tmplID}
	return asynq.NewTask(TypeEmailDelivery, payload)
}

func NewImageResizeTask(src string) *asynq.Task {
	payload := map[string]interface{}{"src": src}
	return asynq.NewTask(TypeImageResize, payload)
}

//---------------------------------------------------------------
// Write a function HandleXXXTask to handle the input task.
// Note that it satisfies the asynq.HandlerFunc interface.
//
// Handler doesn't need to be a function. You can define a type
// that satisfies asynq.Handler interface. See examples below.
//---------------------------------------------------------------

func HandleEmailDeliveryTask(ctx context.Context, t *asynq.Task) error {
	userID, err := t.Payload.GetInt("user_id")
	if err != nil {
		return err
	}
	tmplID, err := t.Payload.GetString("template_id")
	if err != nil {
		return err
	}
	fmt.Printf("Send Email to User: user_id = %d, template_id = %s\n", userID, tmplID)
	// Email delivery code ...
	return nil
}

// ImageProcessor implements asynq.Handler interface.
type ImageProcessor struct {
	// ... fields for struct
}

func (p *ImageProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	src, err := t.Payload.GetString("src")
	if err != nil {
		return err
	}
	fmt.Printf("Resize image: src = %s\n", src)
	// Image resizing code ...
	return nil
}

func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{}
}

const redisAddr = "127.0.0.1:6379"

func main() {
	r := asynq.RedisClientOpt{Addr: redisAddr}
	c := asynq.NewClient(r)
	defer c.Close()

	// ------------------------------------------------------
	// Example 1: Enqueue task to be processed immediately.
	//            Use (*Client).Enqueue method.
	// ------------------------------------------------------

	t := NewEmailDeliveryTask(42, "some:template:id")
	res, err := c.Enqueue(t)
	if err != nil {
		log.Fatalf("could not enqueue task: %v", err)
	}
	fmt.Printf("Enqueued Result: %+v\n", res)


	// ------------------------------------------------------------
	// Example 2: Schedule task to be processed in the future.
	//            Use ProcessIn or ProcessAt option.
	// ------------------------------------------------------------

	t = NewEmailDeliveryTask(42, "other:template:id")
	res, err = c.Enqueue(t, asynq.ProcessIn(24*time.Hour))
	if err != nil {
		log.Fatalf("could not schedule task: %v", err)
	}
	fmt.Printf("Enqueued Result: %+v\n", res)


	// ----------------------------------------------------------------------------
	// Example 3: Set other options to tune task processing behavior.
	//            Options include MaxRetry, Queue, Timeout, Deadline, Unique etc.
	// ----------------------------------------------------------------------------

	c.SetDefaultOptions(TypeImageResize, asynq.MaxRetry(10), asynq.Timeout(3*time.Minute))

	t = NewImageResizeTask("some/blobstore/path")
	res, err = c.Enqueue(t)
	if err != nil {
		log.Fatalf("could not enqueue task: %v", err)
	}
	fmt.Printf("Enqueued Result: %+v\n", res)

	// ---------------------------------------------------------------------------
	// Example 4: Pass options to tune task processing behavior at enqueue time.
	//            Options passed at enqueue time override default ones, if any.
	// ---------------------------------------------------------------------------

	t = NewImageResizeTask("some/blobstore/path")
	res, err = c.Enqueue(t, asynq.Queue("critical"), asynq.Timeout(30*time.Second))
	if err != nil {
		log.Fatalf("could not enqueue task: %v", err)
	}
	fmt.Printf("Enqueued Result: %+v\n", res)

	r = asynq.RedisClientOpt{Addr: redisAddr}

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
	mux.HandleFunc(TypeEmailDelivery, HandleEmailDeliveryTask)
	mux.Handle(TypeImageResize, NewImageProcessor())
	// ...register other handlers...

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}