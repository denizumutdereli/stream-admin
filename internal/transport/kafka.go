package transport

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

type kafkaManager struct {
	logger              *zap.Logger
	brokers             []string
	consumerGroupPrefix string
	maxRetry            int
	retryWait           time.Duration
}

type KafkaManager interface {
	NewConsumer(topics []string, consumerGroup string, autoCommit bool) (*kafka.Consumer, error)
	NewProducer() (*kafka.Producer, error)
}

func NewKafkaManager(brokers []string, consumerGroupPrefix string, maxRetry int, retryWait time.Duration, logger *zap.Logger) (KafkaManager, error) {
	return &kafkaManager{
		logger:              logger,
		brokers:             brokers,
		consumerGroupPrefix: consumerGroupPrefix,
		maxRetry:            maxRetry,
		retryWait:           retryWait,
	}, nil
}

func (k *kafkaManager) NewConsumer(topics []string, consumerGroup string, autoCommit bool) (*kafka.Consumer, error) {
	var consumer *kafka.Consumer
	var err error

	for i := 0; i < k.maxRetry; i++ {
		consumer, err = kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":  k.brokers,
			"group.id":           fmt.Sprintf("%s_%s", k.consumerGroupPrefix, consumerGroup),
			"auto.offset.reset":  "earliest",
			"enable.auto.commit": autoCommit,
		})

		if err != nil {
			fmt.Printf("Failed to create consumer, retry %d/%d\n", i+1, k.maxRetry)
			k.logger.Error("Failed to create consumer", zap.Int("retry", i+1), zap.Int("maxRetry", k.maxRetry))
			time.Sleep(k.retryWait)
		} else {
			break
		}
	}

	if err != nil {
		k.logger.Error("Failed to create consumer after max retries", zap.Int("maxRetry", k.maxRetry), zap.Error(err))
		return nil, fmt.Errorf("failed to create consumer after %d retries: %v", k.maxRetry, err)
	}

	for i := 0; i < k.maxRetry; i++ {
		err = consumer.SubscribeTopics(topics, nil)
		if err != nil {
			k.logger.Error("Failed to subscribe to topics", zap.Int("retry", i+1), zap.Int("maxRetry", k.maxRetry))
			time.Sleep(k.retryWait)
		} else {
			break
		}
	}

	if err != nil {
		k.logger.Error("Failed to subscribe to topics after max retries", zap.Int("maxRetry", k.maxRetry), zap.Error(err))
		return nil, fmt.Errorf("failed to subscribe to topics after %d retries: %v", k.maxRetry, err)
	}

	k.logger.Info("Consumer created and subscribed to topics", zap.Strings("topics", topics))
	return consumer, nil
}

func (k *kafkaManager) NewProducer() (*kafka.Producer, error) {
	var producer *kafka.Producer
	var err error

	for i := 0; i < k.maxRetry; i++ {
		producer, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers": k.brokers,
		})

		if err != nil {
			k.logger.Error("Failed to create producer", zap.Int("retry", i+1), zap.Int("maxRetry", k.maxRetry))
			time.Sleep(k.retryWait)
		} else {
			break
		}
	}

	if err != nil {
		k.logger.Error("Failed to create producer after after max retries", zap.Int("maxRetry", k.maxRetry), zap.Error(err))
		return nil, fmt.Errorf("failed to create producer after %d retries: %v", k.maxRetry, err)
	}

	k.logger.Info("Producer created")
	return producer, nil
}
