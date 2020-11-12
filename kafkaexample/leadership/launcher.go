package leadership

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	"log"
	"time"
)

type LauncherConfig struct {
	brokers       []string
	name          string
	topic         string
	key           string
	groupId       string
	batchInterval int64
}

type Launcher struct {
	cfg           LauncherConfig
	producer      sarama.SyncProducer
	consumer      sarama.ConsumerGroup
	proceedCount  int64
	ctx           context.Context
	cancel        context.CancelFunc
	lastTimestamp int64
}

func (l *Launcher) Start() {
	go l.loopConsumer()
	go l.loopProducer()
}

func (l *Launcher) Close() {
	l.cancel()
	l.consumer.Close()
	l.producer.Close()
}

func (l *Launcher) loopProducer() {
	// produce launch message
	ticker := time.NewTicker(time.Second * 3)
	for {
		select {
		case <-ticker.C:
			b, _ := json.Marshal(&Message{
				LauncherName: l.cfg.name,
				UnixTime:     time.Now().Unix(),
			})
			l.producer.SendMessage(&sarama.ProducerMessage{
				Topic: l.cfg.topic,
				Key:   sarama.StringEncoder(l.cfg.key),
				Value: sarama.StringEncoder(b),
			})
		case <-l.ctx.Done():
			log.Printf("[Producer-%s] Terminate to produce message", l.cfg.name)
			return
		}
	}
}

func (l *Launcher) loopConsumer() {
	// consume launch message to start working
	for {
		err := l.consumer.Consume(l.ctx, []string{l.cfg.topic}, l)
		if l.ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Printf("[Consumer-%s] Error:%v", l.cfg.name, err)
		}
	}
}

//func (l *Launcher) restartConsumer() error {
//	if err := l.setupConsumer(); err != nil {
//		return err
//	}
//	go l.loopConsumer()
//	return nil
//}

func (l *Launcher) Setup(session sarama.ConsumerGroupSession) error {
	log.Printf("[Consumer-%s] Setup is called:%v", l.cfg.name, session)
	return nil
}

func (l *Launcher) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Printf("[Consumer-%s] Cleanup is called:%v", l.cfg.name, session)
	return nil
}

func (l *Launcher) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var m Message
		err := json.Unmarshal(msg.Value, &m)
		if err != nil {
			log.Println("failed to unmarshal message. err:", err)
			return nil
		}
		//log.Printf("[Consume-%s]Consume a message..:%s", l.name, string(msg.Value))
		if l.lastTimestamp+l.cfg.batchInterval < m.UnixTime {
			l.proceedCount++
			log.Printf(">>>>>> [Consume-%s]Launch a work... %d", l.cfg.name, l.proceedCount)
			l.lastTimestamp = m.UnixTime
		}
		if l.proceedCount%5 == 0 {
			log.Printf("TODO: Trigger rebalancing...")
		}
	}
	return nil
}

func (l *Launcher) setupProducer() error {
	pCfg := sarama.NewConfig()
	pCfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(l.cfg.brokers, pCfg)
	if err != nil {
		return nil
	}
	l.producer = producer
	return nil
}

func (l *Launcher) setupConsumer() error {
	cCfg := sarama.NewConfig()
	cCfg.Consumer.Return.Errors = true
	cCfg.Consumer.Offsets.AutoCommit.Enable = false
	cCfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	consumer, err := sarama.NewConsumerGroup(l.cfg.brokers, l.cfg.groupId, cCfg)
	if err != nil {
		return err
	}
	l.consumer = consumer
	return nil
}

func NewLauncher(brokers []string, name, topic, key, groupId string) (*Launcher, error) {
	cfg := LauncherConfig{
		brokers:       brokers,
		name:          name,
		topic:         topic,
		key:           key,
		groupId:       groupId,
		batchInterval: 3,
	}
	l := Launcher{
		cfg: cfg,
	}
	err := l.setupProducer()
	if err != nil {
		return nil, err
	}
	err = l.setupConsumer()
	if err != nil {
		return nil, err
	}
	l.ctx, l.cancel = context.WithCancel(context.Background())

	return &l, nil
}
