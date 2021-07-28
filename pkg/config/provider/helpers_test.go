package provider

import (
	"fmt"
	"testing"

	"github.com/bkrukowski/go-daemon/pkg/config"
	"github.com/stretchr/testify/assert"
)

func Test_cloneStringStringMap(t *testing.T) {
	scenarios := []map[string]string{
		{"foo": "bar", "hello": "world"},
		{"firstname": "Jane", "lastname": "Doe"},
	}

	for id, s := range scenarios {
		t.Run(fmt.Sprintf("scenario #%d", id), func(t *testing.T) {
			cp := cloneStringStringMap(s)
			assert.Equal(t, s, cp)
			assert.NotSame(t, s, cp)
		})
	}
}

func Test_cloneConfig(t *testing.T) {
	scenarios := []config.Config{
		{
			Vars:      map[string]string{"foo": "bar"},
			Templates: make(map[string]string),
			Processes: make(map[string]config.Process),
		},
		{
			Vars: map[string]string{"foo": "bar"},
			Templates: map[string]string{
				"ssh-tunnel": "ssh -N -L {{ .port }}:{{ .localhost }}:{{ .localport }} {{ .ENV_USER }}@{{ .bastionhost }}",
			},
			Processes: make(map[string]config.Process),
		},
	}

	for id, c := range scenarios {
		t.Run(fmt.Sprintf("scenario #%d", id), func(t *testing.T) {
			cp := cloneConfig(c)
			assert.Equal(t, c, cp)
			assert.NotSame(t, c, cp)
		})
	}
}
