package ioutil

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockWriter struct {
	n   int
	err error
}

func (m mockWriter) Write([]byte) (int, error) {
	return m.n, m.err
}

func TestPrefixWriter_Write(t *testing.T) {
	buff := bytes.NewBufferString("")

	scenarios := []struct {
		writer io.Writer
		input  []byte
		prefix string
		n      int
		error  string
		after  func(*testing.T)
	}{
		{
			writer: ioutil.Discard,
			input:  []byte("hello world"),
			prefix: "[info]",
			n:      len("hello world"),
			error:  "",
		},
		{
			writer: mockWriter{n: 0, err: fmt.Errorf("my error")},
			input:  []byte("panic"),
			prefix: "[error]",
			n:      0,
			error:  "my error",
		},
		{
			writer: mockWriter{n: 3 + len("[error]"), err: fmt.Errorf("could not write")},
			input:  []byte("panic"),
			prefix: "[error]",
			n:      3,
			error:  "could not write",
		},
		{
			writer: buff,
			input:  []byte("hello world"),
			prefix: "[info] ",
			n:      len("hello world"),
			error:  "",
			after: func(t *testing.T) {
				assert.Equal(t, "[info] hello world", buff.String())
			},
		},
		{
			writer: mockWriter{n: 0},
			input:  []byte(""),
			n:      len(""),
		},
	}

	for id, s := range scenarios {
		t.Run(fmt.Sprintf("Scenario #%d", id), func(t *testing.T) {
			if s.after != nil {
				defer s.after(t)
			}

			pw := NewPrefixWriter(s.writer, s.prefix)
			n, err := pw.Write(s.input)
			assert.Equal(t, s.n, n)

			if s.error == "" {
				assert.NoError(t, err)
				return
			}

			assert.Errorf(t, err, s.error)
		})
	}
}
