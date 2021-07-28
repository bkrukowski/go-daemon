package provider

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefault(t *testing.T) {
	d := NewDefault(func() string {
		f, err := ioutil.TempFile(os.TempDir(), "go-daemon-test")
		if !assert.NoError(t, err) {
			return ""
		}
		defer func() {
			assert.NoError(t, os.Remove(f.Name()))
		}()
		return f.Name()
	})
	assert.Greater(t, len(d.processors), 0)
}
