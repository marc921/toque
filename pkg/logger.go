package pkg

import "go.uber.org/zap"

func NewLogger(env string, service string) *zap.Logger {
	var logger *zap.Logger
	var err error
	if env == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		zap.L().Fatal("can't initialize zap logger", zap.Error(err), zap.String("service", service))
	}

	return logger.With(zap.String("service", service))
}
