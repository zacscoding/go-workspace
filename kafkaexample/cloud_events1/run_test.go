package cloud_events1

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"testing"
	"time"
)

/*
Kafka message example
1) User
Key: 50dae851-2975-4186-b35e-49b2aa152160
Header: ce_id: 50dae851-2975-4186-b35e-49b2aa152160, ce_source: example.producer, ce_specversion: 1.0, ce_time: 2021-03-12T18:37:52.745553Z, ce_type: user, content-type: application/json
Body: { "id": "user-id-0", "name": "user-name-0"}

2) Product
Key: c00d2b09-da51-482b-ac3c-9957abbc6e0c
Header: ce_id: c00d2b09-da51-482b-ac3c-9957abbc6e0c, ce_source: example.producer, ce_specversion: 1.0, ce_time: 2021-03-12T18:37:55.77472Z, ce_type: product, content-type: application/json
Body: { "id": 3, "price": 103, "stock": 903 }
*/

func Test1(t *testing.T) {
	assert.NoError(t, setupTopic())
	// start consumer
	consumer, err := NewConsumer()
	assert.NoError(t, err)
	defer consumer.Close()
	go func() {
		if err := consumer.Start(); err != nil {
			log.Println("failed to start consumer. err:", err)
		}
	}()

	// start proudcer
	producer, err := NewProducer()
	assert.NoError(t, err)
	for i := 0; i < 10; i++ {
		event := cloudevents.NewEvent()
		event.SetID(uuid.New().String())
		event.SetSource("example.producer")
		if rand.Intn(100)%2 == 0 {
			u := &User{
				ID:   fmt.Sprintf("user-id-%d", i),
				Name: fmt.Sprintf("user-name-%d", i),
			}
			event.SetType(u.Type())
			event.SetData(cloudevents.ApplicationJSON, u)
		} else {
			p := &Product{
				ID:    int64(i),
				Price: 100 + i,
				Stock: 900 + i,
			}
			event.SetType(p.Type())
			event.SetData(cloudevents.ApplicationJSON, p)
		}
		if err := producer.Produce(context.Background(), event); err != nil {
			log.Printf("[#%d] failed to produce event. err:%v", i, err)
		}
		time.Sleep(time.Second)
	}
}

func setupTopic() error {
	admin, err := sarama.NewClusterAdmin(brokers, sarama.NewConfig())
	if err != nil {
		return err
	}
	return admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     int32(numOfPartitions),
		ReplicationFactor: 1,
	}, false)
}
