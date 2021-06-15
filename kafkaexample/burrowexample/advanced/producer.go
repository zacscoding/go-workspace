package main

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"time"
)

type Producer struct {
	Name     string
	Topic    string
	Interval time.Duration

	producer sarama.SyncProducer
	closeCh  chan struct{}
}

func (p *Producer) Start() error {
	cfg := sarama.NewConfig()
	cfg.Version = kafkaVersion
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return err
	}
	p.producer = producer
	p.closeCh = make(chan struct{}, 1)
	go p.loopProduce()
	return nil
}
//1623758923000
//1623758923000

func (p *Producer) Stop() {
	close(p.closeCh)
	p.producer.Close()
}

func (p *Producer) loopProduce() {
	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()
	count := 0

	for {
		select {
		case <-ticker.C:
			m, _ := json.Marshal(&Message{
				Value: fmt.Sprintf("Message-%d", count),
			})
			_, _, err := p.producer.SendMessage(&sarama.ProducerMessage{
				Topic: p.Topic,
				Value: sarama.StringEncoder(m),
			})
			if err != nil {
				log.Println("failed to produce message", err)
			} else {
				count++
			}
		case <-p.closeCh:
			log.Printf("[%s] terminate producer", p.Name)
			return
		}
	}
}
