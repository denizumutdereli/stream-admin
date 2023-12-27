package factory

import (
	"fmt"
	"time"

	"github.com/denizumutdereli/stream-admin/internal/transport"
)

var err error

func (f *serviceFactory) BuildTransports(redis, nats, kafka bool) []error {

	var errors []error

	if redis {
		f.redis, err = f.NewRedisManager()
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to initialize RedisManager: %w", err))
		}
	}

	if nats {
		err = f.NewNatsManager()
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to initialize NatsManager: %w", err))
		}
	}

	if kafka {
		f.kafka, err = f.NewKafkaManager()
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to initialize KafkaManager: %w", err))
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

func (f *serviceFactory) NewNatsManager() error {
	var err error
	f.nats, err = transport.NewNatsManager(f.config.NatsURL, f.logger, f.config)
	if err != nil {
		return fmt.Errorf("failed to initialize Source Nats cluster: %w", err)
	}
	return nil
}

func (f *serviceFactory) NewKafkaManager() (transport.KafkaManager, error) {
	return transport.NewKafkaManager(f.config.KafkaBrokers, f.config.KafkaConsumerGroup, f.config.MaxRetry, time.Duration(f.config.MaxRetry), f.logger)
}

func (f *serviceFactory) NewRedisManager() (*transport.RedisManager, error) {
	return transport.NewRedisManager(f.config.RedisURL, f.config)
}
