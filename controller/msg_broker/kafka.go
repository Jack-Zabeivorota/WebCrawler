package msg_broker

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/segmentio/kafka-go"

	"main/logger"
	"main/models"
	"main/tools"
)

type KafkaBroker struct {
	brokers []string
	writer  *kafka.Writer
	log     *logger.Logger
}

func NewKafkaBroker(hosts string) *KafkaBroker {
	log := logger.Instance()

	var brokers []string

	if hosts != "" {
		brokers = strings.Split(hosts, ",")
	} else {
		brokers = []string{"localhost:9092"}
	}

	var conn *kafka.Conn
	var err error

	for _, br := range brokers {
		tools.RetryCycle(func() error {
			conn, err = kafka.Dial("tcp", br)
			return err
		}, "Kafka fail connecting try", true)

		if err == nil {
			break
		}
	}

	if err != nil {
		log.Fatal("Open connection to Kafka error: %v", err)
	}

	tools.RetryCycle(func() error {
		return conn.CreateTopics(
			kafka.TopicConfig{
				Topic:             models.FindWordsTopic,
				NumPartitions:     3,
				ReplicationFactor: 1,
			},
			kafka.TopicConfig{
				Topic:             models.SignalsTopic,
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
		)
	}, "Kafka fail creating topics try", false)

	err = conn.Close()

	if err != nil {
		log.Error("Close connection to Kafka error: %v", err)
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
	})

	return &KafkaBroker{
		brokers: brokers,
		writer:  writer,
		log:     log,
	}
}

func (kb *KafkaBroker) Send(topic string, messages ...any) error {
	if len(messages) == 0 {
		return nil
	}

	kafkaMessages := tools.Select(messages, func(data any) kafka.Message {
		msg, _ := json.Marshal(data)

		return kafka.Message{
			Topic: topic,
			Value: msg,
		}
	})

	ctx := context.Background()

	err := tools.RetryCycle(func() error {
		return kb.writer.WriteMessages(ctx, kafkaMessages...)
	}, "Kafka fail try", false)

	if err != nil {
		kb.log.Error("Writing messages to Kafka error: %v", err)
	}

	return err
}

func (kb *KafkaBroker) ReceivingWithWaitGroup(wg *sync.WaitGroup, ctx context.Context, topic, group string, handler func([]byte)) error {
	wg.Add(1)
	defer wg.Done()
	return kb.Receiving(ctx, topic, group, handler)
}

func (kb *KafkaBroker) Receiving(ctx context.Context, topic, group string, handler func([]byte)) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: kb.brokers,
		Topic:   topic,
		GroupID: group,
	})
	defer reader.Close()

	kb.log.Info("Start receiving messages from Kafka")
	commit_ctx := context.Background()

	for {
		msg, err := reader.ReadMessage(ctx)

		if err == context.Canceled {
			return nil
		}

		if err != nil {
			kb.log.Error("Reading message form Kafka error: %v", err)
			return err
		}

		handler(msg.Value)

		err = reader.CommitMessages(commit_ctx, msg)

		if err != nil {
			kb.log.Error("Commiting message to Kafka error: %v", err)
			return err
		}
	}
}
