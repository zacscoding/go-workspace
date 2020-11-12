package leadership

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
	"time"
)

var (
	brokers = []string{"localhost:9092"}
	topic   = "launcher-example"
)

type LauncherSuite struct {
	suite.Suite
	admin sarama.ClusterAdmin
}

func (s *LauncherSuite) SetupSuite() {
	// setup admin
	admin, err := sarama.NewClusterAdmin(brokers, sarama.NewConfig())
	s.NoError(err)

	// setup topic
	err = admin.DeleteTopic(topic)
	if err != nil && err != sarama.ErrUnknownTopicOrPartition {
		s.Error(err)
	}
	err = admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     3,
		ReplicationFactor: 1,
	}, false)
	s.NoError(err)
}

func (s *LauncherSuite) TearDownSuite() {
}

func (s *LauncherSuite) TestLauncher() {
	var (
		repeat    = 10
		launchers []*Launcher
	)

	for i := 0; i < repeat; i++ {
		var (
			name    = fmt.Sprintf("launcher-%d", i)
			key     = "launcherkey"
			groupId = "launcher"
		)
		launcher, err := NewLauncher(brokers, name, topic, key, groupId)
		s.NoError(err)
		launcher.Start()
		launchers = append(launchers, launcher)
		time.Sleep(time.Second)
	}

	time.Sleep(time.Minute)

	for _, launcher := range launchers {
		launcher.Close()
	}
}

func (s *LauncherSuite) TestLauncherWithTerminate() {
	var (
		repeat    = 10
		launchers []*Launcher
	)

	for i := 0; i < repeat; i++ {
		var (
			name    = fmt.Sprintf("launcher-%d", i)
			key     = "launcherkey"
			groupId = "launcher"
		)
		launcher, err := NewLauncher(brokers, name, topic, key, groupId)
		s.NoError(err)
		launcher.Start()
		launchers = append(launchers, launcher)
		time.Sleep(time.Second)
	}

	time.Sleep(30 * time.Second)

	for _, launcher := range launchers {
		launcher.Close()
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	}
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(LauncherSuite))
}
