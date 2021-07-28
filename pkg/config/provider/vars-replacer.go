package provider

import (
	"github.com/bkrukowski/go-daemon/pkg/config"
)

type varsReplacer struct {
}

func newVarsReplacer() *varsReplacer {
	return &varsReplacer{}
}

func (v varsReplacer) Process(c config.Config) (config.Config, error) {
	o := cloneConfig(c)

	for _, p := range o.Processes {
		vars := p.Compiled.Vars
		for k, val := range o.Vars {
			vars[k] = val
		}
		for k, val := range p.Vars {
			vars[k] = val
		}
	}

	return o, nil
}
