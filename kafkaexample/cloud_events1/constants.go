package cloud_events1

import (
	"fmt"
	"time"
)

var (
	brokers         = []string{"localhost:9092"}
	topic           = fmt.Sprintf("sample-message-%d", time.Now().Unix())
	numOfPartitions = 1
	groupId         = "sample-consumers"
)
