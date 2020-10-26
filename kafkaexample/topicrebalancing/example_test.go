package topicrebalancing

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTopicRebalancing(t *testing.T) {
	brokers := []string{"localhost:9092"}
	topic := fmt.Sprintf("sample-topic-%d", time.Now().Unix())

	admin, err := sarama.NewClusterAdmin(brokers, sarama.NewConfig())
	assert.NoError(t, err)
	err = admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     3,
		ReplicationFactor: 0,
	}, false)
	assert.NoError(t, err)
	admin.AlterPartitionReassignments()
}
