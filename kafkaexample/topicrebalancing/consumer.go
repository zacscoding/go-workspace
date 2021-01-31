package topicrebalancing

import (
	"context"
	"github.com/Shopify/sarama"
	"log"
)

type MessageConsumer struct {
	name           string
	loggingSession bool
	client         sarama.ConsumerGroup
}

func NewMessageConsumer(name string, loggingSession bool) (*MessageConsumer, error) {
	cfg := sarama.NewConfig()
	//cfg.Version = sarama.MaxVersion
	cfg.Consumer.Offsets.AutoCommit.Enable = false
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	// cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	// cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyPlan{}
	// cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky

	client, err := sarama.NewConsumerGroup(brokers, groupId, cfg)
	if err != nil {
		return nil, err
	}

	consumer := MessageConsumer{
		name:           name,
		loggingSession: loggingSession,
		client:         client,
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
		err := c.client.Consume(ctx, []string{topic}, c)
		if err != nil {
			if err == sarama.ErrClosedClient || err == sarama.ErrClosedConsumerGroup {
				log.Printf("[Consumer-%s] Terminated", c.name)
				return
			}
			log.Printf("[Consumer-%s] failed to consume. err:%v", c.name, err)
		}
	}
}

func (c *MessageConsumer) Setup(session sarama.ConsumerGroupSession) error {
	if c.loggingSession {
		log.Printf("[Consumer-%s] Setup Session. memberid:%s", c.name, session.MemberID())
	}
	return nil
}

func (c *MessageConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	if c.loggingSession {
		log.Printf("[Consumer-%s] Cleanup Session. memberid:%s", c.name, session.MemberID())
	}
	return nil
}

func (c *MessageConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Printf("[Consumer-%s] Consume:%s [partition:%d, offset:%d]", c.name, string(msg.Value), msg.Partition, msg.Offset)
		session.MarkOffset(msg.Topic, msg.Partition, msg.Offset, "")
	}
	return nil
}
