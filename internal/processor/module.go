package processor

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"

	"playGround/pkg/logger"
	"playGround/pkg/tracer"
)

type config struct {
	OrderFilePath            string `envconfig:"PROCESSOR_ORDER_FILE_PATH"`
	AccountStatisticFilePath string `envconfig:"PROCESSOR_ACCOUNT_STATISTIC_FILE_PATH"`
}

func (c *config) Load() error {
	return envconfig.Process("", c)
}

var m *Module

func New(ctx context.Context) (*Module, error) {
	_, span := tracer.Start(ctx, "internal.processor.New")
	defer span.End()

	if m != nil {
		return m, nil
	}

	l := logger.New("processor")

	c := &config{}
	if err := c.Load(); err != nil {
		return nil, errors.Wrap(err, "loading configuration")
	}

	m = &Module{
		config: c,
		log:    l,
	}

	return m, nil
}

type Module struct {
	config *config
	log    logger.Logger
}
