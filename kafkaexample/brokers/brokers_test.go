package brokers

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"log"
	"strconv"
	"testing"
	"time"
)

var (
	brokers = []string{
		"localhost:9092", "localhost:9093", "localhost:9094",
	}
	topic = "sample-message"
)

func TestCreateMessages(t *testing.T) {
	pCfg := sarama.NewConfig()
	pCfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, pCfg)
	assert.NoError(t, err)
	cancelCh := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(time.Millisecond * 100)
		defer ticker.Stop()
		idx := 0
		for {
			select {
			case <-ticker.C:
				partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
					Topic: topic,
					Key:   sarama.StringEncoder(strconv.Itoa(idx)),
					Value: sarama.StringEncoder(fmt.Sprintf("Message-%d", idx)),
				})
				log.Printf("[#%d] partition:%d, offset:%d, err:%s", idx, partition, offset, err)
				idx++
			case <-cancelCh:
				return
			}
		}
	}()
	time.Sleep(5 * time.Minute)
	cancelCh <- struct{}{}
}

func TestCreateTopic(t *testing.T) {
	admin, err := sarama.NewClusterAdmin(brokers, sarama.NewConfig())
	assert.NoError(t, err)
	topics, err := admin.ListTopics()
	assert.NoError(t, err)
	if _, ok := topics[topic]; ok {
		err = admin.DeleteTopic(topic)
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
	}
	err = admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     3,
		ReplicationFactor: 3,
	}, false)
	assert.NoError(t, err)
}