package leadership

import "time"

type BatchType string

const (
	UpperCaseBatchType = "1"
	LowerCaseBatchType = "2"
)

type Message struct {
	LauncherName string
	BatchType    BatchType
}

type Launcher interface {
	StartTrigger(batchType BatchType, interval time.Duration) (<-chan struct{}, error)
	StopTrigger(batchType BatchType, interval time.Duration) (<-chan struct{}, error)
	Close()
}