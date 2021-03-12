package cloud_events1

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"log"
)

type Consumer struct {
	cli      cloudevents.Client
	receiver *kafka_sarama.Consumer
}

func NewConsumer() (*Consumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_0_0_0
	receiver, err := kafka_sarama.NewConsumer(brokers, saramaConfig, groupId, topic)
	if err != nil {
		return nil, err
	}
	cli, err := cloudevents.NewClient(receiver)
	if err != nil {
		receiver.Close(context.Background())
		return nil, err
	}
	consumer := Consumer{
		cli:      cli,
		receiver: receiver,
	}
	return &consumer, nil
}

func (c *Consumer) Start() error {
	return c.cli.StartReceiver(context.Background(), c.onEvent)
}

func (c *Consumer) onEvent(ctx context.Context, event cloudevents.Event) {
	log.Println("Consume message...")
	log.Println("ID:", event.ID())
	log.Println("Source:", event.Source())
	log.Println("Type:", event.Type())
	switch event.Type() {
	case "user":
		var user User
		if err := event.DataAs(&user); err != nil {
			log.Println("failed to data as user. err:", err)
		} else {
			log.Println("User Data():", toJson(&user))
		}
	case "product":
		var product Product
		if err := event.DataAs(&product); err != nil {
			log.Println("failed to data as product. err:", err)
		} else {
			log.Println("User Data():", toJson(&product))
		}
	default:
		log.Println("Unknown data type:", event.Type(), ", data:", string(event.Data()))
	}
}

func (c *Consumer) Close() {
	c.receiver.Close(context.Background())
}

func toJson(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
