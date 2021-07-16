package eventbus

import "github.com/asaskevich/EventBus"

type Producer struct {
	bus EventBus.Bus
}

func (p *Producer) Subscribe() {

}

func (p *Producer) Close() {
	p.bus.WaitAsync()
}
