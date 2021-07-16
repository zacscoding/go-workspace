package main

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	brokers = []string{
		"localhost:9092",
	}
	kafkaVersion = sarama.V2_2_0_0
)

func main() {
	sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
	srv := NewServer()
	go func() {
		if err := srv.Start(":8900"); err != nil {
			log.Fatal(err)
		}
	}()

	exitChannel := make(chan os.Signal, 1)
	signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	<-exitChannel
	log.Println("shutdown")
	srv.Close()
}

type Server struct {
	Producers map[string]*Producer
	Consumers map[string]*Consumer
	e         *echo.Echo
	mutex     sync.Mutex
}

func NewServer() *Server {
	s := &Server{
		Producers: make(map[string]*Producer),
		Consumers: make(map[string]*Consumer),
	}
	s.e = echo.New()
	s.e.POST("/v1/event", s.openEvent)
	s.e.DELETE("/v1/event", s.closeEvent)

	s.e.POST("/v1/producer/:name", s.startProducer)
	s.e.DELETE("/v1/producer/:name", s.stopProducer)
	s.e.POST("/v1/consumer/:name", s.startConsumer)
	s.e.DELETE("/v1/consumer/:name", s.stopConsumer)
	s.e.PUT("/v1/consumer/:name", s.shouldFailConsumer)

	return s
}

func (s *Server) Start(addr string) error {
	return s.e.Start(addr)
}

func (s *Server) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.e.Shutdown(context.Background())
	for _, p := range s.Producers {
		p.Stop()
	}
	for _, c := range s.Consumers {
		c.Stop()
	}
}

func (s *Server) openEvent(c echo.Context) error {
	log.Println("[Notification] Called POST /v1/event")
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.Println("> Failed to read body. err:", err)
		return err
	}
	log.Println("> Body: ", string(body))
	c.JSON(http.StatusOK, nil)
	return nil
}

func (s *Server) closeEvent(c echo.Context) error {
	log.Println("[Notification] Called Delete /v1/event")
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.Println("> Failed to read body. err:", err)
		return err
	}
	log.Println("> Body: ", string(body))
	c.JSON(http.StatusOK, nil)
	return nil
}

func (s *Server) startProducer(c echo.Context) error {
	name, topic, err := extractNameAndTopic(c)
	if err != nil {
		return err
	}
	interval, err := parseDuration(c.QueryParam("interval"))
	if err != nil {
		return err
	}

	id := name + "_" + topic
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.Producers[id]; ok {
		return echo.NewHTTPError(http.StatusConflict, "already exist producer")
	}
	p := Producer{
		Name:     name,
		Topic:    topic,
		Interval: interval,
	}
	if err := p.Start(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	s.Producers[id] = &p
	c.JSON(http.StatusOK, echo.Map{
		"status": "created",
		"metadata": echo.Map{
			"name":     name,
			"topic":    topic,
			"interval": c.QueryParam("interval"),
		},
	})
	return nil
}

func (s *Server) stopProducer(c echo.Context) error {
	name, topic, err := extractNameAndTopic(c)
	if err != nil {
		return err
	}
	id := name + "_" + topic
	s.mutex.Lock()
	defer s.mutex.Unlock()
	p, ok := s.Producers[id]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "producer not found")
	}
	p.Stop()
	delete(s.Producers, id)
	c.JSON(http.StatusOK, echo.Map{
		"status": "deleted",
	})
	return nil
}

func (s *Server) startConsumer(c echo.Context) error {
	name, topic, groupID, err := extractNameTopicGroupID(c)
	if err != nil {
		return err
	}

	id := name + "_" + topic
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.Consumers[id]; ok {
		return echo.NewHTTPError(http.StatusConflict, "already exist consumer")
	}
	consumer := Consumer{
		Name:    name,
		GroupID: groupID,
		Topic:   topic,
	}
	if err := consumer.Start(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	s.Consumers[id] = &consumer
	c.JSON(http.StatusOK, echo.Map{
		"status": "created",
		"metadata": echo.Map{
			"name":    name,
			"topic":   topic,
			"groupId": groupID,
		},
	})
	return nil
}

func (s *Server) stopConsumer(c echo.Context) error {
	name, topic, err := extractNameAndTopic(c)
	if err != nil {
		return err
	}
	id := name + "_" + topic
	s.mutex.Lock()
	defer s.mutex.Unlock()
	consumer, ok := s.Consumers[id]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "consumer not found")
	}
	consumer.Stop()
	delete(s.Consumers, id)
	c.JSON(http.StatusOK, echo.Map{
		"status": "deleted",
	})
	return nil
}

func (s *Server) shouldFailConsumer(c echo.Context) error {
	name, topic, err := extractNameAndTopic(c)
	if err != nil {
		return err
	}
	id := name + "_" + topic
	s.mutex.Lock()
	defer s.mutex.Unlock()
	consumer, ok := s.Consumers[id]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "consumer not found")
	}

	shouldFail := strings.EqualFold("true", c.QueryParam("shouldFail"))
	consumer.SetShouldFail(shouldFail)
	c.JSON(http.StatusOK, echo.Map{
		"shouldFail": shouldFail,
	})
	return nil
}

func extractNameAndTopic(c echo.Context) (name, topic string, err error) {
	name = strings.ToLower(c.Param("name"))
	if name == "" {
		err = echo.NewHTTPError(http.StatusBadRequest, "empty name")
		return
	}
	topic = c.QueryParam("topic")
	if topic == "" {
		err = echo.NewHTTPError(http.StatusBadRequest, "require topic in query params")
		return
	}
	return name, topic, nil
}

func extractNameTopicGroupID(c echo.Context) (name, topic, groupID string, err error) {
	name, topic, err = extractNameAndTopic(c)
	if err != nil {
		return
	}
	groupID = c.QueryParam("groupId")
	if groupID == "" {
		err = echo.NewHTTPError(http.StatusBadRequest, "require groupId in query params")
		return
	}
	return name, topic, groupID, nil
}

func parseDuration(val string) (time.Duration, error) {
	if val == "" {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "require interval in query params")
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid duration: %s", val))
	}
	return d, nil
}
