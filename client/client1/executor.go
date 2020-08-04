package client1

type Executor interface {
	Execute(req func() (interface{}, error)) (interface{}, error)
}

type NoopExecutor struct {
}

func (n *NoopExecutor) Execute(req func() (interface{}, error)) (interface{}, error) {
	return req()
}
