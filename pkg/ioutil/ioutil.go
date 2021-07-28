package ioutil

import (
	"io"
	"sync"
)

type PrefixWriter struct {
	writer    io.Writer
	prefix    string
	addPrefix bool
	locker    sync.Locker
}

func NewPrefixWriter(writer io.Writer, prefix string) *PrefixWriter {
	return &PrefixWriter{writer: writer, prefix: prefix, addPrefix: true, locker: &sync.Mutex{}}
}

func (p *PrefixWriter) Write(b []byte) (n int, err error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	if p.addPrefix {
		n, err = p.withPrefix(b)
	} else {
		n, err = p.writer.Write(b)
	}

	p.addPrefix = len(b) > 0 && string(b[len(b)-1:]) == "\n"

	return
}

func (p *PrefixWriter) withPrefix(b []byte) (n int, err error) {
	n, err = p.writer.Write([]byte(p.prefix + string(b)))
	n -= len(p.prefix)
	if n < 0 {
		n = 0
	}
	return
}
