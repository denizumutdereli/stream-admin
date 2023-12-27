package types

import (
	"github.com/denizumutdereli/stream-admin/internal/caesar"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/transport"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ExchangeConfig struct {
	DB           *gorm.DB
	Nats         *transport.NatsManager
	Redis        *transport.RedisManager
	Websocket    *transport.WSClient
	Kafka        *transport.KafkaManager
	Config       *config.Config
	Logger       *zap.Logger
	Caesar       *caesar.CaesarManager
	Channels     []string
	StreamAssets []string
	Service      *interface{}
}

type ContextualMessage struct {
	UserId                string
	MessageType           string
	Message               string
	RedisDelivery         bool
	RedisTimeoutInMinutes *int
	NatsDelivery          bool
	IssuedAt              int64
}
