package provider

import (
	"os"
	"strings"

	"github.com/bkrukowski/go-daemon/pkg/config"
)

type envAppender struct {
}

func newEnvAppender() *envAppender {
	return &envAppender{}
}

func (e envAppender) Process(c config.Config) (config.Config, error) {
	o := cloneConfig(c)

	for k, v := range e.getAllEnvVars() {
		for _, p := range o.Processes {
			p.Compiled.Vars["ENV_"+k] = v
		}
	}

	return o, nil
}

func (e envAppender) getAllEnvVars() map[string]string {
	r := make(map[string]string)
	for _, v := range os.Environ() {
		k := strings.Split(v, "=")[0]
		r[k] = os.Getenv(k)
	}
	return r
}
