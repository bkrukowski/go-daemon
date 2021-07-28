package provider

import (
	"github.com/bkrukowski/go-daemon/pkg/config"
)

func cloneStringStringMap(i map[string]string) map[string]string {
	o := make(map[string]string)
	for k, v := range i {
		o[k] = v
	}
	return o
}

func cloneConfig(c config.Config) config.Config {
	o := config.Config{
		Vars:      cloneStringStringMap(c.Vars),
		Templates: cloneStringStringMap(c.Templates),
		Processes: make(map[string]config.Process),
	}

	for k, p := range c.Processes {
		np := config.Process{
			Template: p.Template,
			Vars:     cloneStringStringMap(p.Vars),
			Tags:     append(p.Tags),
		}
		np.Compiled.Template = p.Compiled.Template
		np.Compiled.Vars = cloneStringStringMap(p.Compiled.Vars)
		o.Processes[k] = np
	}

	return o
}
