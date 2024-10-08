package messagebroker

import (
	"go.uber.org/zap"
)

type MessageProcessor struct {
	Consumer Consumer
	logger   *zap.Logger
}

func NewMessageProcessor(consumer Consumer, logger *zap.Logger) *MessageProcessor {
	return &MessageProcessor{
		Consumer: consumer,
		logger:   logger,
	}
}
