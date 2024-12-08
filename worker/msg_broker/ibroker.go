package msg_broker

import (
	"context"
	"sync"
)

type IMessageBroker interface {
	Send(topic string, messages ...any) error
	Receiving(ctx context.Context, topic, group string, handler func([]byte)) error
	ReceivingWithWaitGroup(wg *sync.WaitGroup, ctx context.Context, topic, group string, handler func([]byte)) error
}
