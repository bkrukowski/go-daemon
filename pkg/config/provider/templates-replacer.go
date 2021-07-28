package provider

import (
	"github.com/bkrukowski/go-daemon/pkg/config"
)

type templateReplacer struct {
}

func newTemplateReplacer() *templateReplacer {
	return &templateReplacer{}
}

func (t templateReplacer) Process(c config.Config) (config.Config, error) {
	o := cloneConfig(c)

	for n, p := range o.Processes {
		tpl, ok := o.Templates[p.Template]
		if !ok {
			tpl = p.Template
		}

		p.Compiled.Template = tpl
		o.Processes[n] = p
	}

	return o, nil
}
