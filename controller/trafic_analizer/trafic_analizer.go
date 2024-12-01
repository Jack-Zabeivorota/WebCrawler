package trafic_analizer

import (
	"context"
	"sync"
	"time"

	"main/logger"
	"main/models"
	"main/msg_broker"
)

type TraficAnalizer struct {
	msgCounter  int
	lastWorkers int
	mutex       sync.RWMutex
	broker      msg_broker.IMessageBroker
	log         *logger.Logger
	group       string
}

func New(broker msg_broker.IMessageBroker, group string) *TraficAnalizer {
	return &TraficAnalizer{
		broker: broker,
		log:    logger.Instance(),
		group:  group,
	}
}

func (ta *TraficAnalizer) checker() {
	for {
		time.Sleep(time.Minute)
		workers := 0

		if ta.msgCounter < 10 {
			workers = 2
		} else if ta.msgCounter < 50 {
			workers = 5
		} else if ta.msgCounter < 100 {
			workers = 10
		} else if ta.msgCounter < 300 {
			workers = 30
		} else if ta.msgCounter < 600 {
			workers = 60
		} else {
			workers = 100
		}

		if workers != ta.lastWorkers {
			ta.broker.Send(models.SignalsTopic, &models.SignMsg{
				Sign: models.Signals.ScaleUpdate,
				Services: map[string]int{
					"worker": workers,
				},
			})
			ta.lastWorkers = workers
			ta.log.Info("From trafic analizer: messages = %d, workers = %d", ta.msgCounter, workers)
		}

		ta.mutex.Lock()
		ta.msgCounter = 0
		ta.mutex.Unlock()
	}
}

func (ta *TraficAnalizer) handler(data []byte) {
	ta.mutex.Lock()
	ta.msgCounter++
	ta.mutex.Unlock()
}

func (ta *TraficAnalizer) Run() {
	go ta.checker()
	ctx := context.Background()
	ta.broker.Receiving(ctx, models.FindWordsTopic, ta.group, ta.handler)
}
