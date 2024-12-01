package msg_broker

import (
	"context"
	"sync"
)

type MockMessageBroker struct{}

func (MockMessageBroker) Send(topic string, messages ...any) error {
	print("Sent in ", topic, " messages: ", messages, "\n")
	return nil
}

func (MockMessageBroker) Receiving(ctx context.Context, topic, group string, handler func([]byte)) error {
	println("Lisening ", topic, " (group: ", group, ")")
	return nil
}

func (MockMessageBroker) ReceivingWithWaitGroup(wg *sync.WaitGroup, ctx context.Context, topic, group string, handler func([]byte)) error {
	wg.Add(1)
	defer wg.Done()
	println("Lisening ", topic, " (group: ", group, ")")
	return nil
}
