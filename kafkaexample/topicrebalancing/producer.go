package topicrebalancing

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"strconv"
	"time"
)

type MessageProducer struct {
	Producer sarama.SyncProducer
	Proceed  int
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewMessageProducer(interval time.Duration) (*MessageProducer, error) {
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
					Topic:     topic,
					Key:       sarama.StringEncoder(strconv.Itoa(p.Proceed)),
					Value:     sarama.ByteEncoder(bytes),
					Partition: int32(p.Proceed % numOfPartitions),
				})
				p.Proceed++
				//log.Printf("[Producer-P-%d] Success to produce a message:%s", i, message.Value)
			}
		case <-p.ctx.Done():
			log.Println("[Produer] Terminate process.")
			return
		}
	}
}
