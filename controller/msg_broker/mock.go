package msg_broker

import (
	"context"
	"sync"
)

type MockMessageBroker struct{}

func (MockMessageBroker) Send(topic string, messages ...any) error {
	println("Sent in ", topic, " messages: ", messages)
	return nil
}

func (MockMessageBroker) Receiving(ctx context.Context, topic, group string, handler func([]byte)) error {
	println("Lisening ", topic, " (group: ", group, ")")
	<-ctx.Done()
	return ctx.Err()
}

func (MockMessageBroker) ReceivingWithWaitGroup(wg *sync.WaitGroup, ctx context.Context, topic, group string, handler func([]byte)) error {
	wg.Add(1)
	defer wg.Done()
	println("Lisening ", topic, " (group: ", group, ")")
	<-ctx.Done()
	return ctx.Err()
}
