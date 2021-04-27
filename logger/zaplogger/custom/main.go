package main

import (
	"errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
)

var logger *zap.SugaredLogger

func main() {
	if err := initLogger(); err != nil {
		log.Fatal(err)
	}

	logger.Info("info with only messages")
	logger.Infof("info with format :%s", "formatValue")
	logger.Infow("info with key values", "key1", "value1", "key2", "value2")

	if err := callWithError(); err != nil {
		logger.Errorw("error with key values", "err", err)
	}
	// Output
	//{"level_custom_key":"info","time_custom_key":"2021-04-28T08:09:30.394+0900","caller_custom_key":"custom/main.go:17","message_key":"info with only messages"}
	//{"level_custom_key":"info","time_custom_key":"2021-04-28T08:09:30.395+0900","caller_custom_key":"custom/main.go:18","message_key":"info with format :formatValue"}
	//{"level_custom_key":"info","time_custom_key":"2021-04-28T08:09:30.395+0900","caller_custom_key":"custom/main.go:19","message_key":"info with key values","key1":"value1","key2":"value2"}
	//{"level_custom_key":"error","time_custom_key":"2021-04-28T08:09:30.395+0900","caller_custom_key":"custom/main.go:22","message_key":"error with key values","err":"custom error","strack_trace_key":"main.main\n\t/workspace/git/zaccoding/go-workspace/logger/zaplogger/custom/main.go:22\nruntime.main\n\t/usr/local/go/src/runtime/proc.go:203"}
}

func callWithError() error {
	return errors.New("custom error")
}

func initLogger() error {
	ec := zapcore.EncoderConfig{
		TimeKey:        "time_custom_key",
		LevelKey:       "level_custom_key",
		NameKey:        "name_custom_key",
		CallerKey:      "caller_custom_key",
		MessageKey:     "message_key",
		StacktraceKey:  "strack_trace_key",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	cfg := zap.Config{
		Encoding:         "json",
		EncoderConfig:    ec,
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:      false,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	l, err := cfg.Build()
	if err != nil {
		return err
	}
	logger = l.Sugar()
	return nil
}
