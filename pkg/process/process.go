package process

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

const (
	NotDone = math.MinInt32

	SIGKILLDelay = time.Second * 5
)

type Option func(*process)

func WithStdout(w io.Writer) Option {
	return func(p *process) {
		p.stdout = w
	}
}

func WithStderr(w io.Writer) Option {
	return func(p *process) {
		p.stderr = w
	}
}

type process struct {
	name   string
	args   []string
	stdout io.Writer
	stderr io.Writer

	cmd  *exec.Cmd
	done chan struct{}

	err      error
	exitCode int

	locker sync.Locker
}

func New(name string, args []string, opts ...Option) *process {
	r := &process{
		name:     name,
		args:     args,
		exitCode: NotDone,
		locker:   &sync.Mutex{},
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Start starts the given process.
// p := New("echo", []string{"hello", "world"})
// err := p.Start()
// if err != nil {
//     return err
// }
// <-p.Done()
func (p *process) Start(ctx context.Context) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	if p.cmd != nil {
		return fmt.Errorf("process has been started already")
	}

	p.cmd = exec.Command(p.name, p.args...)
	p.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	p.cmd.Stdout = p.stdout
	p.cmd.Stderr = p.stderr

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("could not start process: %w", err)
	}

	pgid, pgidErr := syscall.Getpgid(p.cmd.Process.Pid)

	p.done = make(chan struct{})

	// Do not use "os/exec".Cmd.Process.Kill()
	// see https://stackoverflow.com/questions/22470193/why-wont-go-kill-a-child-process-correctly

	// Once the context is done, kill the process by sending the following signals:
	// * syscall.SIGTERM
	// * syscall.SIGKILL
	// with a delay till the process is done.
	go func() {
		select {
		case <-p.done:
		// do nothing
		case <-ctx.Done():
			if pgidErr != nil {
				_ = p.cmd.Process.Kill()
				return
			}
			_ = syscall.Kill(-pgid, syscall.SIGTERM)
			t := time.NewTimer(SIGKILLDelay)
			defer t.Stop()

			select {
			case <-p.done:
				return
			case <-t.C:
				_ = syscall.Kill(-pgid, syscall.SIGKILL)
			}
		}
	}()

	go func() {
		err := p.cmd.Wait()

		ec := 0
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				ec = exitErr.ExitCode()
				err = nil
			}
		}
		p.exitCode = ec
		p.err = err
		close(p.done)
	}()

	return nil
}

// Error returns error returned by exec.Cmd{}.Wait().
// In case of exec.ExitError, Error returns nil
// and exit code will be returned by ExitCode.
func (p *process) Error() error {
	return p.err
}

// ExitCode returns exit code or NotDone when the process hasn't been done yet.
func (p *process) ExitCode() int {
	return p.exitCode
}

// Done returns a channel that's closed when the process is done.
// Before calling Done, Start must be called, otherwise it returns nil.
func (p *process) Done() <-chan struct{} {
	return p.done
}
