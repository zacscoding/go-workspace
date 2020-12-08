package legacy

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"go-workspace/logger/zaplogger"
	"log"
	"testing"
	"time"
)

var (
	brokers = []string{"localhost:9092"}
	topic   = "launcher-example"
)

func TestKafkaLauncherWithConsole(t *testing.T) {
	setupTopic(t)
	var (
		count     = 10
		launchers []Launcher
	)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("launcher-%d", i)
		logger := zaplogger.DefaultLogger().With("launcher", name)
		cfg := KafkaLauncherConfig{
			LauncherName: name,
			Brokers:      brokers,
			KafkaCfg:     NewDefaultKafkaConfig(),
			Topic:        topic,
			GroupId:      "launcher-consumer",
			MessageKey:   "launcherKey",
			Interval:     5 * time.Second,
		}
		launcher, err := NewKafkaLauncher(logger, &cfg)
		assert.NoError(t, err)
		launchers = append(launchers, launcher)
		go func(l Launcher) {
			triggerChan := l.Start()
			for {
				select {
				case <-triggerChan:
					//sleep := rand.Intn(5)
					sleep := 0
					log.Printf("working %s... will sleep %d secs", name, sleep)
					//time.Sleep(time.Second * time.Duration(sleep))
				}
			}
		}(launcher)
	}
	time.Sleep(time.Minute)
	for _, launcher := range launchers {
		launcher.Close()
		time.Sleep(30 * time.Second)
	}
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
