package processdef

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateProcessFromTemplate(t *testing.T) {
	scenarios := []struct {
		tpl    string
		params []string
		err    string
	}{
		{
			tpl:    "echo hello world",
			params: []string{"echo", "hello", "world"},
		},
		{
			tpl: " ",
			err: "invalid command",
		},
		{
			tpl:    "pkill      -f pattern",
			params: []string{"pkill", "-f", "pattern"},
		},
		{
			tpl: "echo 'hello world",
			err: "invalid command: EOF found when expecting closing quote",
		},
	}

	for id, s := range scenarios {
		t.Run(fmt.Sprintf("scenario #%d", id), func(t *testing.T) {
			p, err := CreateProcessFromTemplate("foo", nil, s.tpl)
			if s.err != "" {
				assert.EqualError(t, err, s.err)
				assert.Zero(t, p)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, s.params, append([]string{p.Name}, p.Args...))
		})
	}
}
