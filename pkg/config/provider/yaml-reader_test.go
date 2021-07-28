package provider

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/bkrukowski/go-daemon/pkg/config"
	"github.com/stretchr/testify/assert"
)

func Test_yamlReader_Process(t *testing.T) {
	const contents = `
vars:
  ENV_USER: "jane"
  bastionhost: "bastionhost.local"

templates:
  ssh-tunnel: "ssh -N -L {{ .port }}:{{ .localhost }}:{{ .localport }} {{ .ENV_USER }}@{{ .bastionhost }}"

processes:
  clock:
    template: "bash -c 'while sleep {{ .sleep }}; do date +\"%T\"; done'"
    vars:
      sleep: 1
    tags: ["demo"]

  es:
    template: ssh-tunnel
    vars:
      port: 9201
      localhost: elasticsearch.local
      localport: 9200
`

	t.Run("Given scenario", func(t *testing.T) {
		file, err := ioutil.TempFile(os.TempDir(), "go-daemon-test")
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			assert.NoError(t, os.Remove(file.Name()))
		}()

		_, err = file.Write([]byte(contents))
		if !assert.NoError(t, err) {
			return
		}

		r := newYamlReader(func() string {
			return file.Name()
		})
		c, err := r.Process(config.Config{})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(
			t,
			config.Config{
				Vars: map[string]string{
					"ENV_USER":    "jane",
					"bastionhost": "bastionhost.local",
				},
				Templates: map[string]string{
					"ssh-tunnel": "ssh -N -L {{ .port }}:{{ .localhost }}:{{ .localport }} {{ .ENV_USER }}@{{ .bastionhost }}",
				},
				Processes: map[string]config.Process{
					"clock": {
						Template: "bash -c 'while sleep {{ .sleep }}; do date +\"%T\"; done'",
						Vars:     map[string]string{"sleep": "1"},
						Tags:     []string{"demo"},
					},
					"es": {
						Template: "ssh-tunnel",
						Vars: map[string]string{
							"port":      "9201",
							"localhost": "elasticsearch.local",
							"localport": "9200",
						},
					},
				},
			},
			c,
		)
	})

	t.Run("File does not exist", func(t *testing.T) {
		file, err := ioutil.TempFile(os.TempDir(), "go-daemon-test")
		if !assert.NoError(t, err) {
			return
		}
		if !assert.NoError(t, os.Remove(file.Name())) {
			return
		}

		fmt.Println(file.Name())
		r := newYamlReader(func() string {
			return file.Name()
		})
		c, err := r.Process(config.Config{})
		assert.EqualError(t, err, fmt.Sprintf("could not read file `%s`: open %s: no such file or directory", file.Name(), file.Name()))
		assert.Zero(t, c)
	})

	t.Run("Invalid yaml", func(t *testing.T) {
		file, err := ioutil.TempFile(os.TempDir(), "go-daemon-test")
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			assert.NoError(t, os.Remove(file.Name()))
		}()

		_, err = file.Write([]byte("yaml"))
		if !assert.NoError(t, err) {
			return
		}

		r := newYamlReader(func() string {
			return file.Name()
		})
		c, err := r.Process(config.Config{})
		assert.EqualError(t, err, fmt.Sprintf("could not parse yml `%s`: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `yaml` into config.Config", file.Name()))
		assert.Zero(t, c)
	})
}
