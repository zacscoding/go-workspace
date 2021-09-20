package dlock

import "log"

type Logger interface {
	Infof(msg string, args ...interface{})
	Debugf(msg string, args ...interface{})
	Error(msg string)
}

type StdLogger struct{}

func (s *StdLogger) Infof(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

func (s *StdLogger) Debugf(msg string, args ...interface{}) {
	log.Printf("[DEBUG] "+msg, args...)
}

func (s *StdLogger) Error(msg string) {
	log.Printf("[ERROR] " + msg)
}

type NoopLogger struct{}

func (n *NoopLogger) Infof(string, ...interface{})  {}
func (n *NoopLogger) Debugf(string, ...interface{}) {}
func (n *NoopLogger) Error(string)                  {}
