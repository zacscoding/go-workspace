package cloud_events1

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type Producer struct {
	sender *kafka_sarama.Sender
	cli    cloudevents.Client
}

func NewProducer() (*Producer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_0_0_0
	sender, err := kafka_sarama.NewSender(brokers, saramaConfig, topic)
	if err != nil {
		return nil, err
	}
	cli, err := cloudevents.NewClient(sender, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		return nil, err
	}
	p := Producer{
		sender: sender,
		cli:    cli,
	}
	return &p, nil
}

func (p *Producer) Produce(ctx context.Context, e cloudevents.Event) error {
	if result := p.cli.Send(
		kafka_sarama.WithMessageKey(ctx, sarama.StringEncoder(e.ID())),
		e); cloudevents.IsUndelivered(result) {
		return result
	}
	return nil
}

func (p *Producer) Close() {
	p.sender.Close(context.Background())
}
