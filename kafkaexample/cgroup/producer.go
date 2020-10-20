package cgroup

type Producer struct {
	sequence int
}

//func (p *Producer) produceLoop() {
//	for {
//		messages := []Message{
//			NewMessage(p.sequence),
//			NewMessage(p.sequence + 1),
//			NewMessage(p.sequence + 2),
//		}
//		p.sequence += 3
//
//		time.Sleep(time.Second)
//	}
//}
