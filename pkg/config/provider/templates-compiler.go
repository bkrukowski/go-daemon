package provider

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/bkrukowski/go-daemon/pkg/config"
)

type templatesCompiler struct {
}

func newTemplatesCompiler() *templatesCompiler {
	return &templatesCompiler{}
}

func (t templatesCompiler) Process(c config.Config) (config.Config, error) {
	o := cloneConfig(c)

	for n, p := range o.Processes {
		tpl, err := template.New(n).
			Option("missingkey=error").
			Parse(p.Compiled.Template)

		if err != nil {
			return config.Config{}, fmt.Errorf("could not parse template `%s`: %s", n, err.Error())
		}

		var b bytes.Buffer
		err = tpl.Execute(&b, p.Compiled.Vars)
		if err != nil {
			return config.Config{}, fmt.Errorf("could not execute template `%s`: %s", n, err.Error())
		}

		p.Compiled.Template = b.String()
		o.Processes[n] = p
	}

	return o, nil
}
