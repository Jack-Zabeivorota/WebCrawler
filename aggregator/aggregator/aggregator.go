package aggregator

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"main/cache"
	"main/database"
	"main/logger"
	"main/models"
	"main/msg_broker"
)

type Aggregator struct {
	id                   int
	waitGroup            *sync.WaitGroup
	cancelTasksReceiving context.CancelFunc
	cancelSignsReceiving context.CancelFunc
	taskRecepients       int
	broker               msg_broker.IMessageBroker
	cache                cache.ICache
	db                   database.IDataBase
	log                  *logger.Logger
	group                string
}

func New() *Aggregator {
	log := logger.Instance()
	db := database.NewPostgreSQL(os.Getenv("POSTGRES_CONN"))
	err := db.Init()

	if err != nil {
		log.Fatal("Initialization DB error: %v", err)
	}

	id, err := strconv.Atoi(os.Getenv("ID"))

	if err != nil {
		log.Fatal("ID is not specified")
	}

	taskRecepients, err := strconv.Atoi(os.Getenv("TASK_RECEPIENTS"))

	if err != nil {
		taskRecepients = 1
	}

	return &Aggregator{
		id:             id,
		waitGroup:      &sync.WaitGroup{},
		taskRecepients: taskRecepients,
		broker:         msg_broker.NewKafkaBroker(os.Getenv("KAFKA_HOSTS")),
		cache:          cache.NewRedisCache(os.Getenv("REDIS_HOST")),
		db:             db,
		log:            log,
		group:          "Aggregators",
	}
}

func (agg *Aggregator) runTaskRecipients(ctx context.Context) {
	agg.waitGroup.Add(agg.taskRecepients)

	for i := 0; i < agg.taskRecepients; i++ {
		go agg.broker.ReceivingWithWaitGroup(agg.waitGroup, ctx, models.AggregateResultTopic, agg.group, agg.handler)
	}
}

func (agg *Aggregator) signsHandler(data []byte) {
	msg := &models.SignMsg{}
	err := json.Unmarshal(data, &msg)

	if err != nil {
		agg.log.Error("Incorrect signal error: %v", err)
		return
	}

	if id, ok := msg.Services["aggregator"]; ok {

		if msg.Sign == models.Signals.Shutdown {
			if id == agg.id {
				agg.cancelTasksReceiving()
				agg.cancelTasksReceiving = nil
				agg.cancelSignsReceiving()
				agg.log.Info("Closed with signal 'Shutdown'")
			}
			return
		}

		if msg.Sign == models.Signals.ScaleUpdate {

			if agg.id <= id && agg.cancelTasksReceiving == nil {
				ctx := context.Background()
				ctx, agg.cancelTasksReceiving = context.WithCancel(ctx)

				agg.runTaskRecipients(ctx)
				agg.log.Info("Start with signal 'ScaleUpdate'")

			} else if agg.id > id && agg.cancelTasksReceiving != nil {
				agg.cancelTasksReceiving()
				agg.cancelTasksReceiving = nil
				agg.log.Info("Pause with signal 'ScaleUpdate'")
			}
			return

		}

		if msg.Sign == models.Signals.Kill && id == agg.id {
			agg.log.Info("Closed with signal 'Kill'")
			os.Exit(0)
		}

	}
}

func (agg *Aggregator) Run() {
	ar_ctx, s_ctx := context.Background(), context.Background()
	ar_ctx, agg.cancelTasksReceiving = context.WithCancel(ar_ctx)
	s_ctx, agg.cancelSignsReceiving = context.WithCancel(s_ctx)

	agg.waitGroup.Add(1)
	go agg.broker.ReceivingWithWaitGroup(agg.waitGroup, s_ctx, models.SignalsTopic, agg.group, agg.signsHandler)
	agg.runTaskRecipients(ar_ctx)

	agg.waitGroup.Wait()
}

func (agg *Aggregator) handler(data []byte) {
	agg.log.Info("Start 'handler'")

	// parsing message and checking completed urls

	msg := &models.AggregateResultMsg{}
	err := json.Unmarshal(data, msg)

	if err != nil {
		agg.log.Error("Incorrect message. data = %v", data)
		return
	}

	isCompleted, err := agg.cache.AllURLsIsCompleted(msg.RequestID)

	if err != nil || !isCompleted {
		agg.log.Info("Completed 'handler' for request %d: Fake call", msg.RequestID)
		return
	}

	// aggregating urls results and saving them to DB

	urlResults, err := agg.cache.GetURLsResult(msg.RequestID)

	if err != nil {
		return
	}

	err = agg.db.AddRequestResults(msg.RequestID, urlResults)

	if err != nil {
		return
	}

	agg.cache.ClearRequest(msg.RequestID)

	agg.log.Info("Completed 'handler' for request %d: Done!", msg.RequestID)
}
