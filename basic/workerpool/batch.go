package workerpool

import "gopkg.in/go-playground/pool.v3"

type Batch interface {
	Queue(fn pool.WorkFunc)

	QueueComplete()

	Cancel()

	Results() <-chan pool.WorkUnit

	WaitAll()
}

type batch struct {
}

func (b *batch) Queue(fn pool.WorkFunc) {
	panic("implement me")
}

func (b *batch) QueueComplete() {
	panic("implement me")
}

func (b *batch) Cancel() {
	panic("implement me")
}

func (b *batch) Results() <-chan pool.WorkUnit {
	panic("implement me")
}

func (b *batch) WaitAll() {
	panic("implement me")
}
