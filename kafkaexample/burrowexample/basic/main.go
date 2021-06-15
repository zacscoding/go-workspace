package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	brokers = []string{
		"localhost:9092",
	}
	kafkaVersion = sarama.V2_2_0_0
	topic        = "test-topic"
	groupId      = "consumers-1"
)

type Message struct {
	Value string `json:"value"`
}

func main() {
	var (
		ctx, cancel = context.WithCancel(context.Background())
		group       = sync.WaitGroup{}
	)

	go startNotificationServer(":8900")
	go loopConsumer(ctx, &group, "c1", groupId, false)
	go loopConsumer(ctx, &group, "c2", groupId, true)
	go loopProducer(ctx, &group)

	exitChannel := make(chan os.Signal, 1)
	signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	<-exitChannel
	log.Println("shutdown")
	cancel()
	log.Println("wait group workers..")
	group.Wait()
	log.Println("terminate")
}

func startNotificationServer(addr string) {
	e := echo.New()
	e.POST("/v1/event", func(c echo.Context) error {
		log.Println("[Notification] Called POST /v1/event")
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			log.Println("> Failed to read body. err:", err)
			return err
		}
		log.Println("> Body: ", string(body))
		c.JSON(http.StatusOK, nil)
		return nil
	})
	e.DELETE("/v1/event", func(c echo.Context) error {
		log.Println("[Notification] Called Delete /v1/event")
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			log.Println("> Failed to read body. err:", err)
			return err
		}
		log.Println("> Body: ", string(body))
		c.JSON(http.StatusOK, nil)
		return nil
	})
	if err := e.Start(addr); err != nil {
		log.Fatal(err)
	}
}

func loopProducer(ctx context.Context, group *sync.WaitGroup) {
	group.Add(1)
	defer group.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	cfg := sarama.NewConfig()
	cfg.Version = kafkaVersion
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()
	count := 0

	for {
		select {
		case <-ticker.C:
			m, _ := json.Marshal(&Message{
				Value: fmt.Sprintf("Message-%d", count),
			})
			_, _, err := producer.SendMessage(&sarama.ProducerMessage{
				Topic: topic,
				Value: sarama.StringEncoder(m),
			})
			if err != nil {
				log.Println("failed to produce message", err)
			} else {
				count++
			}
		case <-ctx.Done():
			log.Println("terminate producer")
			return
		}
	}
}

func loopConsumer(ctx context.Context, group *sync.WaitGroup, name, groupId string, shouldFail bool) {
	group.Add(1)
	defer group.Done()

	c := Consumer{
		name:       name,
		shouldFail: shouldFail,
	}

	cfg := sarama.NewConfig()
	cfg.Version = kafkaVersion
	cfg.ClientID = "my-kafka"
	cfg.Consumer.Group.Session.Timeout = time.Second * 6
	cfg.Consumer.Group.Heartbeat.Interval = time.Second
	cfg.Consumer.Offsets.AutoCommit.Enable = false
	cfg.Consumer.Return.Errors = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky

	client, err := sarama.NewConsumerGroup(brokers, groupId, cfg)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		select {
		case <-ctx.Done():
			client.Close()
		}
	}()
	for {
		err := client.Consume(ctx, []string{topic}, &c)
		if ctx.Err() != nil {
			log.Printf("Consumer-%s terminate consumer", c.name)
			return
		}
		if err != nil {
			log.Printf("Consumer-%s failed to consume. err: %v", c.name, err)
		} else {
			log.Printf("Consumer-%s no error after consume", c.name)
		}
	}
}

type Consumer struct {
	name       string
	shouldFail bool
	proceed    int
}

func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	log.Printf("[Consumer-%s] Setup is called:%v", c.name, session)
	return nil
}

func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Printf("[Consumer-%s]Cleanup is called:%v", c.name, session)
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		c.proceed++
		if c.proceed%50 == 0 {
			log.Printf("[Consumer-%s] consume message proceed: %d", c.name, c.proceed)
		}
		if c.shouldFail && c.proceed%100 != 0 {
			continue
		} else {
			session.MarkMessage(msg, "")
			session.Commit()
		}
	}
	return nil
}
