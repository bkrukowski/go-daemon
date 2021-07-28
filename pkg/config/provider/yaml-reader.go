package provider

import (
	"fmt"
	"io/ioutil"

	"github.com/bkrukowski/go-daemon/pkg/config"
	"gopkg.in/yaml.v2"
)

type yamlReader struct {
	fileFinder func() string
}

func newYamlReader(fileFinder func() string) *yamlReader {
	return &yamlReader{fileFinder: fileFinder}
}

func (y yamlReader) Process(_ config.Config) (config.Config, error) {
	fn := y.fileFinder()

	f, err := ioutil.ReadFile(fn)

	if err != nil {
		return config.Config{}, fmt.Errorf("could not read file `%s`: %s", fn, err.Error())
	}

	r := config.Config{}

	err = yaml.Unmarshal(f, &r)
	if err != nil {
		return config.Config{}, fmt.Errorf("could not parse yml `%s`: %s", fn, err.Error())
	}

	return r, nil
}
