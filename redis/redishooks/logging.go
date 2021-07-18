package redishooks

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
)

type LoggingHookParams struct {
	BeforeProcess         bool
	AfterProcess          bool
	BeforeProcessPipeline bool
	AfterProcessPipeline  bool
}

type LoggingHook struct {
	beforeProcess         bool
	afterProcess          bool
	beforeProcessPipeline bool
	afterProcessPipeline  bool
}

func NewLoggingHook(p LoggingHookParams) *LoggingHook {
	return &LoggingHook{
		beforeProcess:         p.BeforeProcess,
		afterProcess:          p.AfterProcess,
		beforeProcessPipeline: p.BeforeProcessPipeline,
		afterProcessPipeline:  p.AfterProcessPipeline,
	}
}

func (l *LoggingHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	if l.beforeProcess {
		log.Printf("[Redis] BeforeProcess: %v", cmd)
	}
	return ctx, nil
}

func (l *LoggingHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if l.afterProcess {
		log.Printf("[Redis] AfterProcess: %v", cmd)
	}
	return nil
}

func (l *LoggingHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	if l.beforeProcessPipeline {
		log.Printf("[Redis] BeforeProcessPipeline: %v", cmds)
	}
	return ctx, nil
}

func (l *LoggingHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if l.afterProcessPipeline {
		log.Printf("[Redis] AfterProcessPipeline: %v", cmds)
	}
	return nil
}
