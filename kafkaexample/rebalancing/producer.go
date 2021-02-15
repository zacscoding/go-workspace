package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"strconv"
	"time"
)

type Message struct {
	Value string `json:"value"`
}

type MessageProducer struct {
	Producer sarama.SyncProducer
	Topic    string
	Proceed  int
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewMessageProducer(brokers []string, topic string, interval time.Duration) (*MessageProducer, error) {
	pcfg := sarama.NewConfig()
	pcfg.Producer.Return.Successes = true
	pcfg.Producer.Partitioner = sarama.NewManualPartitioner
	pcfg.Version = sarama.MaxVersion
	producer, err := sarama.NewSyncProducer(brokers, pcfg)
	if err != nil {
		return nil, err
	}
	p := MessageProducer{
		Producer: producer,
		Proceed:  1,
		Topic:    topic,
	}
	go p.loopProduce(interval)
	return &p, nil
}

func (p *MessageProducer) Stop() {
	p.cancel()
}

func (p *MessageProducer) loopProduce(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	p.ctx, p.cancel = context.WithCancel(context.Background())
	for {
		select {
		case <-ticker.C:
			for i := 0; i < numOfPartitions; i++ {
				message := Message{Value: fmt.Sprintf("Value-%d", p.Proceed)}
				bytes, _ := json.Marshal(message)
				p.Producer.SendMessage(&sarama.ProducerMessage{
					Topic:     p.Topic,
					Key:       sarama.StringEncoder(strconv.Itoa(p.Proceed)),
					Value:     sarama.ByteEncoder(bytes),
					Partition: int32(p.Proceed % numOfPartitions),
				})
				p.Proceed++
			}
		case <-p.ctx.Done():
			log.Println("[Produer] Terminate process.")
			return
		}
	}
}
