package daemon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Option func(*daemon)

type OnProcessDone func(id string, exitCode int) error

func WithOnProcessDone(e OnProcessDone) Option {
	return func(d *daemon) {
		d.onProcessDone = e
	}
}

func WithDelay(d time.Duration) Option {
	return func(de *daemon) {
		de.delay = d
	}
}

// WithIgnoreExitCode option skips checking the exit code.
// By default, daemon returns an error when the process returns a non-zero exit code.
func WithIgnoreExitCode() Option {
	return func(d *daemon) {
		d.ignoreExitCode = true
	}
}

type daemon struct {
	id             string
	factory        processFactory
	delay          time.Duration
	started        bool
	ignoreExitCode bool
	locker         sync.Locker
	onProcessDone  OnProcessDone
}

func New(id string, factory processFactory, opts ...Option) *daemon {
	r := &daemon{
		id:             id,
		factory:        factory,
		delay:          0,
		started:        false,
		ignoreExitCode: false,
		locker:         &sync.Mutex{},
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

type Process interface {
	Start(context.Context) error
	Done() <-chan struct{}
	ExitCode() int
	Error() error
}

type processFactory func() Process

const (
	errPrefix = "daemon{}.Daemonize(): id `%s`: "
	errTwice  = "cannot start the same daemon twice"
)

// Daemonize runs the given Process in the background.
// Whenever the Process gets finished, it will be re-triggered.
func (d *daemon) Daemonize(ctx context.Context) error {
	errorf := func(f string, i ...interface{}) error {
		return fmt.Errorf(errPrefix+f, append([]interface{}{d.id}, i...)...)
	}

	d.locker.Lock()
	if d.started {
		return errorf(errTwice)
	}
	d.started = true
	d.locker.Unlock()

	for {
		p := d.factory()
		if err := p.Start(ctx); err != nil {
			if errors.Is(err, ctx.Err()) {
				return nil
			}
			return errorf("%w", err)
		}

		getErr := func() error {
			if err := p.Error(); err != nil {
				return errorf("unexpected error: %w", err)
			}

			if !d.ignoreExitCode {
				if ec := p.ExitCode(); ec != 0 {
					return errorf("exit code: %d", ec)
				}
			}

			return nil
		}

		select {
		case <-ctx.Done():
			// The given process uses the same context.
			// We can assume, once the context is done,
			// we can wait till the process is done
			// because the process should be notified by context about the necessity to end.
			<-p.Done()
			return getErr()
		case <-p.Done():
			if d.onProcessDone != nil {
				if err := d.onProcessDone(d.id, p.ExitCode()); err != nil {
					return errorf("unexpected error when the event OnProcessDone has been triggered: %w", err)
				}
			}
			if err := getErr(); err != nil {
				return err
			}
			time.Sleep(d.delay)
			continue
		}
	}
}
