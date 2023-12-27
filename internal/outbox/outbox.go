package outbox

import (
	"encoding/json"
	"log"
	"time"

	"github.com/denizumutdereli/stream-admin/internal/config"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OutboxMessage struct {
	ID        uint   `gorm:"primaryKey"`
	Payload   string `gorm:"type:text"`
	State     string `gorm:"type:varchar(20)"`
	Retry     int    `gorm:"type:int"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	StatePending   = "pending"
	StatePublished = "published"
	StateFailed    = "failed"
	MaxRetries     = 5
)

type DispatcherSettings struct {
	ProcessInterval           time.Duration
	LockCheckerInterval       time.Duration
	CleanupWorkerInterval     time.Duration
	MaxLockTimeDuration       time.Duration
	MessagesRetentionDuration time.Duration
}

func DefaultDispatcherSettings() DispatcherSettings {
	return DispatcherSettings{
		ProcessInterval:           20 * time.Second,
		LockCheckerInterval:       600 * time.Minute,
		CleanupWorkerInterval:     60 * time.Second,
		MaxLockTimeDuration:       5 * time.Minute,
		MessagesRetentionDuration: 1 * time.Minute,
	}
}

type OutboxManager struct {
	config      *config.Config
	logger      *zap.Logger
	db          *gorm.DB
	settings    DispatcherSettings
	outboxTable string
}

func NewOutboxManager(config *config.Config, db *gorm.DB, settings DispatcherSettings, outboxTable string) *OutboxManager {
	return &OutboxManager{
		config:      config,
		logger:      config.Logger,
		db:          db,
		settings:    settings,
		outboxTable: outboxTable,
	}
}

func (manager *OutboxManager) CreateMessage(tx *gorm.DB, payload interface{}) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	message := OutboxMessage{
		Payload: string(bytes),
		State:   StatePending,
		Retry:   0,
	}

	return tx.Table(manager.outboxTable).Create(&message).Error
}

func (manager *OutboxManager) ProcessMessages() {
	processTicker := time.NewTicker(manager.settings.ProcessInterval)
	cleanupTicker := time.NewTicker(manager.settings.CleanupWorkerInterval)

	defer processTicker.Stop()
	defer cleanupTicker.Stop()

	for {
		select {
		case <-processTicker.C:
			manager.processPendingMessages()
		case <-cleanupTicker.C:
			manager.cleanupOldMessages()
		}
	}
}

func (manager *OutboxManager) processPendingMessages() {
	var messages []OutboxMessage
	if err := manager.db.Table(manager.outboxTable).Where("state = ?", StatePending).Find(&messages).Error; err != nil {
		log.Printf("Failed to fetch pending messages: %v", err)
		return
	}

	for _, msg := range messages {
		if err := manager.publishMessage(&msg); err != nil {
			manager.handleFailure(&msg, err)
			continue
		}
		manager.markAsPublished(&msg)
	}
}

func (manager *OutboxManager) cleanupOldMessages() {
	expirationTime := time.Now().Add(-manager.settings.MessagesRetentionDuration)
	if err := manager.db.Table(manager.outboxTable).Where("created_at < ? AND state = ?", expirationTime, StatePublished).Delete(&OutboxMessage{}).Error; err != nil {
		log.Printf("Failed to clean up old messages: %v", err)
	}
}

func (manager *OutboxManager) publishMessage(msg *OutboxMessage) error {
	// Simulation TODO: cdc + sink
	log.Printf("Publishing message: %s", msg.Payload)
	return nil
}

func (manager *OutboxManager) handleFailure(msg *OutboxMessage, err error) {
	msg.Retry++
	if msg.Retry > MaxRetries || time.Since(msg.CreatedAt) > manager.settings.MaxLockTimeDuration {
		msg.State = StateFailed
	} else {
		msg.State = StatePending
	}
	manager.db.Table(manager.outboxTable).Save(msg)
}

func (manager *OutboxManager) markAsPublished(msg *OutboxMessage) {
	msg.State = StatePublished
	manager.db.Table(manager.outboxTable).Save(msg)
}
