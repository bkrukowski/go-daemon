package provider

import (
	"fmt"

	"github.com/bkrukowski/go-daemon/pkg/config"
)

type configProcessor interface {
	Process(config config.Config) (config.Config, error)
}

type ConfigProvider struct {
	processors []configProcessor
}

func New(processors []configProcessor) *ConfigProvider {
	return &ConfigProvider{processors: processors}
}

func (cp ConfigProvider) Provide() (c config.Config, err error) {
	for _, p := range cp.processors {
		c, err = p.Process(c)
		if err != nil {
			err = fmt.Errorf("could not provide configuration: %s", err.Error())
			return
		}
	}

	return c, err
}
