package legacy

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
	"time"
)

type Message struct {
	LauncherName string
	UnixTime     int64
}

type Launcher interface {
	Start() <-chan struct{}
	Close()
}

type KafkaLauncherConfig struct {
	LauncherName string
	Brokers      []string
	KafkaCfg     *sarama.Config
	Topic        string
	GroupId      string
	MessageKey   string
	Interval     time.Duration
}

type kafkaLauncher struct {
	logger          *zap.SugaredLogger
	cfg             *KafkaLauncherConfig
	ctx             context.Context
	cancel          context.CancelFunc
	producer        sarama.SyncProducer
	consumer        sarama.ConsumerGroup
	triggerChan     chan struct{}
	nextTriggerTime time.Time
	lastTriggerTime int64
}

func (kl *kafkaLauncher) Close() {
	kl.cancel()
	kl.producer.Close()
	kl.consumer.Close()
}

func (kl *kafkaLauncher) Start() <-chan struct{} {
	kl.ctx, kl.cancel = context.WithCancel(context.Background())
	go kl.loopConsumer()
	go kl.loopProducer()
	return kl.triggerChan
}

// NewDefaultKafkaConfig returns a default kafka config
func NewDefaultKafkaConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	// producer configs
	cfg.Producer.Return.Successes = true
	// consumer configs
	cfg.Consumer.Return.Errors = true
	cfg.Consumer.Offsets.AutoCommit.Enable = false
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	return cfg
}

func NewKafkaLauncher(logger *zap.SugaredLogger, cfg *KafkaLauncherConfig) (Launcher, error) {
	kl := kafkaLauncher{
		logger:      logger,
		cfg:         cfg,
		triggerChan: make(chan struct{}),
	}
	if err := kl.setup(); err != nil {
		return nil, err
	}
	return &kl, nil
}

// setup initializes kafka producer and consumer
func (kl *kafkaLauncher) setup() error {
	producer, err := sarama.NewSyncProducer(kl.cfg.Brokers, kl.cfg.KafkaCfg)
	if err != nil {
		return err
	}
	consumer, err := sarama.NewConsumerGroup(kl.cfg.Brokers, kl.cfg.GroupId, kl.cfg.KafkaCfg)
	if err != nil {
		return err
	}
	kl.producer = producer
	kl.consumer = consumer
	return nil
}

// loopProducer produce launch kafka messages with given internal / 10
func (kl *kafkaLauncher) loopProducer() {
	var (
		nextProduce = time.Now().Add(kl.cfg.Interval)
		ticker      = time.NewTicker(kl.cfg.Interval / 10)
	)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			if now.After(nextProduce) {
				bytes, _ := json.Marshal(&Message{
					LauncherName: kl.cfg.LauncherName,
					UnixTime:     time.Now().Unix(),
				})
				kl.producer.SendMessage(&sarama.ProducerMessage{
					Topic: kl.cfg.Topic,
					Key:   sarama.StringEncoder(kl.cfg.MessageKey),
					Value: sarama.StringEncoder(bytes),
				})
				nextProduce = now.Add(kl.cfg.Interval)
			}
		case <-kl.ctx.Done():
			kl.logger.Debugf("[KafkaLauncher-%s] terminate producer", kl.cfg.LauncherName)
			return
		}
	}
}

func (kl *kafkaLauncher) loopConsumer() {
	for {
		err := kl.consumer.Consume(kl.ctx, []string{kl.cfg.Topic}, kl)
		if kl.ctx.Err() != nil {
			kl.logger.Debugf("[KafkaLauncher-%s] terminate consumer", kl.cfg.LauncherName)
			return
		}
		if err != nil {
			kl.logger.Errorw("failed to consume launch message", "topic", kl.cfg.Topic, "err", err)
		}
	}
}

func (kl *kafkaLauncher) Setup(session sarama.ConsumerGroupSession) error {
	// TODO : update taking leadership
	kl.logger.Infof("[Launcher-%s] take ownership", kl.cfg.LauncherName)
	return nil
}

func (kl *kafkaLauncher) Cleanup(session sarama.ConsumerGroupSession) error {
	kl.logger.Infof("[Launcher-%s] loose ownership", kl.cfg.LauncherName)
	return nil
}

func (kl *kafkaLauncher) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var m Message
		err := json.Unmarshal(msg.Value, &m)
		if err != nil {
			kl.logger.Errorw("failed to unmarshal launch message", "err", err)
			return nil
		}
		// skip to do
		if time.Now().Before(kl.nextTriggerTime) {
			continue
		}
		kl.triggerChan <- struct{}{}
		kl.nextTriggerTime = time.Now().Add(kl.cfg.Interval)
	}
	return nil
}
