package main

import (
	"context"
	"github.com/Shopify/sarama"
	"log"
	"sync"
)

type MessageConsumer struct {
	name      string
	memberId  string
	groupId   string
	brokers   []string
	topics    []string
	client    sarama.ConsumerGroup
	mutex     sync.Mutex
	ownership map[string]map[int32]struct{}
}

func NewMessageConsumer(name, groupId string, brokers, topics []string, cfg *sarama.Config) (*MessageConsumer, error) {
	client, err := sarama.NewConsumerGroup(brokers, groupId, cfg)
	if err != nil {
		return nil, err
	}
	consumer := MessageConsumer{
		name:    name,
		client:  client,
		groupId: groupId,
		brokers: brokers,
		topics:  topics,
	}
	go consumer.loopConsume()
	return &consumer, nil
}

func (c *MessageConsumer) Stop() {
	c.client.Close()
}

func (c *MessageConsumer) loopConsume() {
	ctx := context.Background()
	for {
		err := c.client.Consume(ctx, c.topics, c)
		if err != nil {
			if err == sarama.ErrClosedClient || err == sarama.ErrClosedConsumerGroup {
				log.Printf("[Consumer-%s] Terminated", c.name)
				return
			}
			log.Printf("[Consumer-%s] failed to consume. err:%v", c.name, err)
		}
	}
}

func (c *MessageConsumer) GetMetadata() map[string]interface{} {
	return map[string]interface{}{
		"name":      c.name,
		"memberId":  c.memberId,
		"ownership": c.ownership,
	}
}

func (c *MessageConsumer) Setup(session sarama.ConsumerGroupSession) error {
	log.Printf("[Consumer-%s] Setup Session. memberid:%s", c.name, session.MemberID())
	c.ownership = make(map[string]map[int32]struct{})
	c.memberId = session.MemberID()
	return nil
}

func (c *MessageConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Printf("[Consumer-%s] Cleanup Session. memberid:%s", c.name, session.MemberID())
	c.ownership = make(map[string]map[int32]struct{})
	return nil
}

func (c *MessageConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		partitions, ok := c.ownership[msg.Topic]
		if !ok {
			partitions = make(map[int32]struct{})
			c.ownership[msg.Topic] = partitions
		}
		c.mutex.Lock()
		partitions[msg.Partition] = struct{}{}
		c.mutex.Unlock()
		log.Printf("[Consumer-%s] Consume:%s [partition:%d, offset:%d]", c.name, string(msg.Value), msg.Partition, msg.Offset)
		session.MarkOffset(msg.Topic, msg.Partition, msg.Offset, "")
	}
	return nil
}
