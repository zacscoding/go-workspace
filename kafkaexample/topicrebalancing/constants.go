package topicrebalancing

import (
	"fmt"
	"time"
)

var (
	brokers         = []string{"localhost:9092"}
	topic           = fmt.Sprintf("sample-message-%d", time.Now().Unix())
	numOfPartitions = 3
	groupId         = "sample-consumers"
)

type Message struct {
	Value string `json:"value"`
}
