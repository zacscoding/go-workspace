package main

import (
	"context"
	"github.com/Shopify/sarama"
	"log"
	"time"
)

type Consumer struct {
	Name    string
	GroupID string
	Topic   string

	client     sarama.ConsumerGroup
	proceed    int
	shouldFail bool
}

func (c *Consumer) Start() error {
	cfg := sarama.NewConfig()
	cfg.Version = kafkaVersion
	cfg.ClientID = "my-kafka"
	cfg.Consumer.Group.Session.Timeout = time.Second * 6
	cfg.Consumer.Group.Heartbeat.Interval = time.Second
	cfg.Consumer.Offsets.AutoCommit.Enable = false
	cfg.Consumer.Return.Errors = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	client, err := sarama.NewConsumerGroup(brokers, c.GroupID, cfg)
	if err != nil {
		return err
	}
	c.client = client
	go c.loopConsume()
	return nil
}

func (c *Consumer) Stop() {
	c.client.Close()
}

func (c *Consumer) SetShouldFail(shouldFail bool) {
	c.shouldFail = shouldFail
}

func (c *Consumer) loopConsume() {
	ctx := context.Background()
	for {
		err := c.client.Consume(ctx, []string{c.Topic}, c)
		if sarama.ErrClosedClient == err {
			log.Printf("[Consumer-%s] terminate", c.Name)
			return
		}
		if err != nil {
			log.Printf("Consumer-%s failed to consume. err: %v", c.Name, err)
		} else {
			log.Printf("Consumer-%s no error after consume", c.Name)
		}
	}
}

func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	log.Printf("[Consumer-%s] Setup is called:%v", c.Name, session)
	return nil
}

func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Printf("[Consumer-%s]Cleanup is called:%v", c.Name, session)
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		c.proceed++
		if c.proceed%50 == 0 {
			log.Printf("[Consumer-%s] consume message proceed: %d", c.Name, c.proceed)
		}
		if c.shouldFail {
			continue
		} else {
			session.MarkMessage(msg, "")
			session.Commit()
		}
	}
	return nil
}
