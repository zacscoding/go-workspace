package leadership

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

var (
	brokers = []string{"localhost:9092"}
	topic   = "launcher-example"
)

func TestTemp(t *testing.T) {
	setupTopic(t)

	cfg := NewDefaultKafkaConfig()
	kafkaCli, err := sarama.NewClient(brokers, cfg)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	kafkaProducer, err := sarama.NewSyncProducerFromClient(kafkaCli)
	assert.NoError(t, err)
	p := MessageProducer{
		interval:    time.Second,
		ctx:         ctx,
		topic:       topic,
		messageType: "2",
		producer:    kafkaProducer,
	}
	go p.loopProducer()

	kafkaConsumer, err := sarama.NewConsumerGroupFromClient("consumer-1", kafkaCli)
	assert.NoError(t, err)
	c := MessageConsumer{
		ctx:      ctx,
		consumer: kafkaConsumer,
		topics:   []string{topic},
	}
	go c.loopConsumer()

	time.Sleep(time.Minute)
	fmt.Println("Try to close kafka cli")
	kafkaCli.Close()
	fmt.Println("Wait 10 secs")
	time.Sleep(10 * time.Second)
	fmt.Println("Try to cancel")
	cancel()
	fmt.Println("Wait 10 secs")
	time.Sleep(10 * time.Second)
}

func setupTopic(t *testing.T) {
	// setup admin
	admin, err := sarama.NewClusterAdmin(brokers, sarama.NewConfig())
	assert.NoError(t, err)

	// setup topic
	err = admin.DeleteTopic(topic)
	if err != nil && err != sarama.ErrUnknownTopicOrPartition {
		assert.NoError(t, err)
	}

	err = admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     3,
		ReplicationFactor: 1,
	}, false)
	fmt.Println("Create a topic :", err)
}

type SampleMessage struct {
	MessageType string `json:"messageType"`
}

type MessageProducer struct {
	interval    time.Duration
	ctx         context.Context
	topic       string
	messageType string
	producer    sarama.SyncProducer
}

func (p *MessageProducer) loopProducer() {
	var (
		nextProduce = time.Now().Add(p.interval)
		ticker      = time.NewTicker(p.interval / 2)
	)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if time.Now().After(nextProduce) {
				bytes, _ := json.Marshal(&SampleMessage{
					MessageType: p.messageType,
				})
				partition, offset, err := p.producer.SendMessage(&sarama.ProducerMessage{
					Topic: p.topic,
					Key:   sarama.StringEncoder(p.messageType),
					Value: sarama.StringEncoder(bytes),
				})
				log.Printf("[Producer] send a message. partition:%d, offset:%d, err:%v",
					partition, offset, err)
				if err == sarama.ErrClosedClient {
					log.Println("[Producer] terminate producer loop because ErrClosedClient")
					return
				}
			}
		case <-p.ctx.Done():
			log.Println("[Producer] terminate producer loop")
			return
		}
	}
}

func NewDefaultKafkaConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	// producer configs
	cfg.Producer.Return.Successes = true
	// consumer configs
	cfg.Consumer.Return.Errors = true
	cfg.Consumer.Offsets.AutoCommit.Enable = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	return cfg
}

type MessageConsumer struct {
	ctx      context.Context
	consumer sarama.ConsumerGroup
	topics   []string
}

func (c *MessageConsumer) loopConsumer() {
	for {
		err := c.consumer.Consume(c.ctx, c.topics, c)
		if c.ctx.Err() != nil {
			log.Println("[Consumer] cancelled consumer")
			return
		}
		if err != nil {
			log.Printf("[Consumer] Error occur:%v", err)
			if err == sarama.ErrClosedClient {
				log.Println("[Consumer] terminate producer loop because ErrClosedClient")
				return
			}
		}
	}
}

func (m *MessageConsumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (m *MessageConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (m *MessageConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Printf("[Consumer] Consumer message: %s", string(msg.Value))
	}
	return nil
}
