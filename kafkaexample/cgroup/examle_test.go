package cgroup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestTopic(t *testing.T) {
	brokers := []string{"localhost:9092"}
	topic := fmt.Sprintf("sample-message-%d", time.Now().Unix())
	admin, err := sarama.NewClusterAdmin(brokers, sarama.NewConfig())
	assert.NoError(t, err)
	err = admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     10,
		ReplicationFactor: 1,
	}, false)

	fmt.Println("First create topic:", err)

	err = admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     10,
		ReplicationFactor: 3,
	}, false)
	fmt.Println("Second create topic:", err)
}

func TestConsumerGroup(t *testing.T) {
	brokers := []string{"localhost:9092"}
	topic := fmt.Sprintf("sample-message-%d", time.Now().Unix())
	groupId := fmt.Sprintf("group-%d", time.Now().Unix())

	// setup topics by admin
	admin, err := sarama.NewClusterAdmin(brokers, sarama.NewConfig())
	assert.NoError(t, err)
	err = admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     10,
		ReplicationFactor: 3,
	}, false)
	assert.NoError(t, err)

	// setup producer
	pCfg := sarama.NewConfig()
	pCfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, pCfg)
	assert.NoError(t, err)

	// setup consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer1 := Consumer{name: "consumer-1", ready: make(chan bool)}
	consumer2 := Consumer{name: "consumer-2", ready: make(chan bool)}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		client, err := sarama.NewConsumerGroup(brokers, groupId, newConsumerConfig())
		assert.NoError(t, err)
		for {
			err := client.Consume(ctx, []string{topic}, &consumer1)
			assert.NoError(t, err)
			if ctx.Err() != nil {
				return
			}
			consumer1.ready = make(chan bool)
		}
	}()
	go func() {
		defer wg.Done()
		client, err := sarama.NewConsumerGroup(brokers, groupId, newConsumerConfig())
		assert.NoError(t, err)
		for {
			err := client.Consume(ctx, []string{topic}, &consumer2)
			assert.NoError(t, err)
			if ctx.Err() != nil {
				return
			}
			consumer2.ready = make(chan bool)
		}
	}()
	<-consumer1.ready
	<-consumer2.ready

	// produce messages
	for i := 0; i < 100; i++ {
		producer.SendMessage(&sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(strconv.Itoa(i)),
			Value: sarama.StringEncoder(fmt.Sprintf("Message-%d", i)),
		})
		time.Sleep(time.Second)
	}
}

func newConsumerConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.MaxVersion
	cfg.Consumer.Offsets.AutoCommit.Enable = true
	cfg.Consumer.Return.Errors = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	return cfg
}

type Consumer struct {
	name  string
	ready chan bool
}

func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	fmt.Printf("[Consumer-%s] Setup is called:%v\n", c.name, session)
	close(c.ready)
	return nil
}

func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	fmt.Printf("[Consumer-%s]Cleanup is called:%v\n", c.name, session)
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		m := make(map[string]interface{})
		headers := make(map[string]interface{})
		for _, header := range msg.Headers {
			headers[string(header.Key)] = string(header.Value)
		}
		m["headers"] = headers
		m["metadata"] = map[string]interface{}{
			"topic":     msg.Topic,
			"partition": msg.Partition,
			"offset":    msg.Offset,
		}
		m["key"] = string(msg.Key)
		m["message"] = string(msg.Value)
		b, _ := json.Marshal(m)
		markMessage := rand.Intn(5) != 0
		fmt.Printf("[Consumer-%s-Mark:%v]%s\n", c.name, markMessage, string(b))
		if markMessage {
			session.MarkMessage(msg, "")
		} else {
			return errors.New("force err")
		}
	}
	return nil
}
