package fetchque

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test1(t *testing.T) {
	bus := EventBus.New()

	fmt.Println("1")
	err := bus.SubscribeOnce("topic1:type1:key1", func(value string) {
		fmt.Println("Subscribe1-1>> topic1:type1:key1:", value)
	})
	assert.NoError(t, err)
	err = bus.SubscribeOnceAsync("topic1:type1:key1", func(value string) {
		fmt.Println("Subscribe1-2>> topic1:type1:key1:", value)
	})
	assert.NoError(t, err)

	fmt.Println("2")
	err = bus.SubscribeOnce("topic1:type1:key2", func(value string) {
		fmt.Println("Subscribe2-1>> topic1:type1:key2:", value)
		time.Sleep(time.Second * 5)
	})
	err = bus.SubscribeOnceAsync("topic1:type1:key2", func(value string) {
		fmt.Println("Subscribe2-2>> topic1:type1:key2:", value)
	})
	assert.NoError(t, err)
	fmt.Println("3")
	bus.Publish("topic1:type1:key1", "value1")
	bus.Publish("topic1:type1:key2", "value2")
	time.Sleep(time.Second * 5)
	bus.Publish("topic1:type1:key1", "value1")
	bus.Publish("topic1:type1:key2", "value2")
}

func Test2(t *testing.T) {
	bus := EventBus.New()
	err := bus.SubscribeOnceAsync("topic1:type1:key1", func(value string) {
		fmt.Println("Subscribe1-1>> topic1:type1:key1:", value)
		time.Sleep(5 * time.Second)
	})
	assert.NoError(t, err)
	bus.Publish("topic1:type1:key1", "value1")
	bus.Publish("topic1:type1:key1", "value2")
	bus.WaitAsync()
	fmt.Println("Completed")
	err = bus.SubscribeOnce("topic1:type1:key1", func(value string) {
		fmt.Println("Subscribe1-1>> topic1:type1:key1:", value)
	})

}
