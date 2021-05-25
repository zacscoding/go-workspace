package basic

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sort"
	"strconv"
	"testing"
	"time"
)

var (
	brokers = []string{"localhost:9092"}
	topic   = "sample-message"
)

func TestTopics(t *testing.T) {
	topic = "sample-topic-" + uuid.New().String()
	admin, err := sarama.NewClusterAdmin(brokers, sarama.NewConfig())
	assert.NoError(t, err)
	err = admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     10,
		ReplicationFactor: 1,
	}, false)
	assert.NoError(t, err)

	err = admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     10,
		ReplicationFactor: 1,
	}, false)

	if terr, ok := err.(*sarama.TopicError); ok {
		if terr.Err == sarama.ErrTopicAlreadyExists {
			fmt.Println("sarama.ErrTopicAlreadyExists")
		} else {
			fmt.Println("sarama.ErrTopicAlreadyExists is not")
		}
	}
	fmt.Println("ERR: ", err)
}

func TestConsumer(t *testing.T) {

	// setup topics by admin
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
		NumPartitions:     10,
		ReplicationFactor: 3,
	}, false)
	assert.NoError(t, err)

	// setup producer & produce messages
	pCfg := sarama.NewConfig()
	pCfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, pCfg)
	assert.NoError(t, err)
	maxOffset := int64(0)
	type message struct {
		partition int32
		offset    int64
	}
	var messages []message

	for i := 0; i < 10; i++ {
		partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(strconv.Itoa(i)),
			Value: sarama.StringEncoder(fmt.Sprintf("Message-%d", i)),
		})
		assert.NoError(t, err)
		if maxOffset < offset {
			maxOffset = offset
		}
		messages = append(messages, message{
			partition: partition,
			offset:    offset,
		})
	}

	// read messages from offset
	cCfg := sarama.NewConfig()
	cCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumer, err := sarama.NewConsumer(brokers, cCfg)
	assert.NoError(t, err)
	sort.Slice(messages, func(i, j int) bool {
		return rand.Intn(10)%2 == 0
	})

	for _, m := range messages {
		c, err := consumer.ConsumePartition(topic, m.partition, m.offset)
		assert.NoError(t, err)
		m2 := <-c.Messages()
		fmt.Println("Consume:" + string(m2.Value))
		c.Close()
	}
	cli, err := sarama.NewClient(brokers, sarama.NewConfig())
	assert.NoError(t, err)
	oldestOffset, err := cli.GetOffset(topic, messages[0].partition, sarama.OffsetOldest)
	assert.NoError(t, err)
	newestOffset, err := cli.GetOffset(topic, messages[0].partition, sarama.OffsetNewest)
	assert.NoError(t, err)
	fmt.Printf(">OldestOffset:%d, NewestOffset:%d\n", oldestOffset, newestOffset)
}
