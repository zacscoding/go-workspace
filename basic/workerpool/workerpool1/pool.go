package workerpool1

type TaskFunc func() (interface{}, error)

type workerPool1 struct {
	limited    uint
	workerChan chan struct{}
	jobChan    chan *TaskFunc
	cancelChan chan struct{}
}

func (p *workerPool1) loopWorker() {
	for {
		select {
		case job := <-p.jobChan:
			p.workerChan <- struct{}{}
		case <-p.cancelChan:
			return
		}
	}
}

func NewWorkerPool1(limited uint) *workerPool1 {
	w := workerPool1{
		limited:    limited,
		workerChan: make(chan struct{}, limited),
		cancelChan: make(chan struct{}),
	}
	go w.loopWorker()
	return &w
}
