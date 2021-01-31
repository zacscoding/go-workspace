package topicrebalancing

import (
	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"log"
	"strconv"
	"testing"
	"time"
)

func TestPartitionAssignator(t *testing.T) {
	assert.NoError(t, setupTopic())
	var (
		consumerId           = 1
		initialConsumers     []*MessageConsumer
		initialConsumerCount = 2
		dynamicConsumers     []*MessageConsumer
		dynamicConsumerCount = 10
	)
	log.Printf("## Try to start initial consumers:#%d", initialConsumerCount)
	for i := 0; i < initialConsumerCount; i++ {
		c, err := NewMessageConsumer(strconv.Itoa(consumerId), true)
		assert.NoError(t, err)
		initialConsumers = append(initialConsumers, c)
		consumerId++
	}
	p, err := NewMessageProducer(time.Second)
	assert.NoError(t, err)
	time.Sleep(time.Second * 15)

	log.Printf("## Try to start dynamic consumers:#%d", dynamicConsumerCount)
	for i := 0; i < dynamicConsumerCount; i++ {
		c, err := NewMessageConsumer(strconv.Itoa(i), false)
		assert.NoError(t, err)
		dynamicConsumers = append(dynamicConsumers, c)
		consumerId++
	}
	time.Sleep(time.Second * 10)

	log.Println("## Try to stop dynamic consumers")
	for _, consumer := range dynamicConsumers {
		consumer.Stop()
		time.Sleep(time.Second)
	}

	for _, c := range initialConsumers {
		log.Printf("## Try to stop: consumer-%s", c.name)
		c.Stop()
		time.Sleep(time.Second)
	}
	p.Stop()
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
