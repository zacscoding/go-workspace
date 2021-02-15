package main

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

var (
	brokers []string
	topics  []string
)

const (
	numOfPartitions = 3
	groupId         = "local-test"
)

func init() {
	brokers = []string{"localhost:9092"}
	now := time.Now().Unix()
	topics = []string{
		fmt.Sprintf("sample-message-%d", now),
		fmt.Sprintf("sample-message-%d", now+1),
	}
}

func main() {
	if err := setupTopic(); err != nil {
		log.Fatal(err)
	}
	var producers []*MessageProducer
	for _, topic := range topics {
		p, err := NewMessageProducer(brokers, topic, time.Second)
		if err != nil {
			log.Fatal(err)
		}
		producers = append(producers, p)
	}
	defer func() {
		for _, p := range producers {
			p.Stop()
		}
	}()
	consumers := make(map[string]*MessageConsumer)
	e := gin.Default()
	e.POST("/:name/start", func(ctx *gin.Context) {
		name, err := bindNameFromURI(ctx)
		if err != nil {
			return
		}
		if _, ok := consumers[name]; ok {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "already exist consumer:" + name,
			})
			return
		}
		c, err := NewMessageConsumer(name, groupId, brokers, topics, newConsumerConfigs())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
		consumers[c.name] = c
	})
	e.POST("/:name/stop", func(ctx *gin.Context) {
		name, err := bindNameFromURI(ctx)
		if err != nil {
			return
		}
		c, ok := consumers[name]
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "not exist consumer:" + name,
			})
			return
		}
		c.Stop()
		delete(consumers, name)
		ctx.JSON(http.StatusOK, gin.H{
			"message": "success to stop consumer:" + name,
		})
	})
	e.GET("/", func(ctx *gin.Context) {
		var ret []map[string]interface{}
		for _, consumer := range consumers {
			ret = append(ret, consumer.GetMetadata())
		}
		sort.Slice(ret, func(i, j int) bool {
			return strings.Compare(ret[i]["memberId"].(string), ret[j]["memberId"].(string)) < 0
		})

		ctx.JSON(http.StatusOK, ret)
	})

	if err := e.Run(":8880"); err != nil {
		log.Fatal(err)
	}
}

func newConsumerConfigs() *sarama.Config {
	cfg := sarama.NewConfig()
	//cfg.Version = sarama.MaxVersion
	cfg.Consumer.Offsets.AutoCommit.Enable = false
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	return cfg
}

func bindNameFromURI(ctx *gin.Context) (string, error) {
	type Uri struct {
		Name string `uri:"name"`
	}
	var uri Uri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return "", err
	}
	return uri.Name, nil
}

func setupTopic() error {
	admin, err := sarama.NewClusterAdmin(brokers, sarama.NewConfig())
	if err != nil {
		return err
	}
	for _, topic := range topics {
		if err := admin.CreateTopic(topic, &sarama.TopicDetail{
			NumPartitions:     int32(numOfPartitions),
			ReplicationFactor: 1,
		}, false); err != nil {
			return err
		}
	}
	return nil
}
