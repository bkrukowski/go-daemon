package daemon

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

type processMock struct {
	error      error
	startError error
	done       <-chan struct{}
	exitCode   int
}

func (p processMock) Error() error {
	return p.error
}

func (p processMock) Start(context.Context) error {
	return p.startError
}

func (p processMock) Done() <-chan struct{} {
	return p.done
}

func (p processMock) ExitCode() int {
	return p.exitCode
}

func Test_daemon_Daemonize(t *testing.T) {
	t.Run("Call Daemonize twice", func(t *testing.T) {
		eg, ctx := errgroup.WithContext(context.Background())
		// It will simulate that process is finished once context has been canceled.
		p := &processMock{done: ctx.Done()}
		d := New("my-process", func() Process {
			return p
		})

		for i := 0; i < 2; i++ {
			eg.Go(func() error {
				return d.Daemonize(ctx)
			})
		}

		assert.EqualError(
			t,
			eg.Wait(),
			"daemon{}.Daemonize(): id `my-process`: cannot start the same daemon twice",
		)
	})
}
