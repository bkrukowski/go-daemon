package runner

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/bkrukowski/go-daemon/pkg/daemon"
	"github.com/bkrukowski/go-daemon/pkg/daemons"
	pkgIoutil "github.com/bkrukowski/go-daemon/pkg/ioutil"
	"github.com/bkrukowski/go-daemon/pkg/process"
	"github.com/bkrukowski/go-daemon/pkg/processdef"
)

type Request struct {
	Names          []string
	Tags           []string
	IgnoreExitCode bool
	Verbose        bool
}

type Service struct {
	processes []processdef.Process
	out       io.Writer
	err       io.Writer
}

func NewService(processes []processdef.Process, out io.Writer, err io.Writer) *Service {
	return &Service{processes: processes, out: out, err: err}
}

func (s *Service) Run(ctx context.Context, r Request) error {
	if len(s.processes) == 0 {
		return fmt.Errorf("at least one process must be defined")
	}

	processes, err := s.filter(r)
	if err != nil {
		return err
	}

	if err := s.printHeader(processes); err != nil {
		return fmt.Errorf("could not print header: %w", err)
	}

	return s.run(ctx, processes, r)
}

func (s *Service) run(ctx context.Context, processes []processdef.Process, r Request) error {
	opts := []daemon.Option{daemon.WithDelay(time.Second)}
	if r.IgnoreExitCode {
		opts = append(opts, daemon.WithIgnoreExitCode())
	}
	if r.Verbose {
		opts = append(opts, daemon.WithOnProcessDone(func(id string, exitCode int) (err error) {
			_, err = fmt.Fprintf(s.out, "[%s] exit code %d\n", id, exitCode)
			return
		}))
	}

	var dmns []daemons.Daemon
	for _, tmp := range processes {
		p := tmp
		stdout, stderr := s.getStdOutErr(p.ID, r.Verbose)
		d := daemon.New(
			p.ID,
			func() daemon.Process {
				return process.New(
					p.Name,
					p.Args,
					process.WithStdout(stdout),
					process.WithStderr(stderr),
				)
			},
			opts...,
		)

		dmns = append(dmns, d)
	}
	return daemons.New(dmns).Daemonize(ctx)
}

func (s *Service) printHeader(processes []processdef.Process) error {
	max := 0
	for _, p := range processes {
		if len(p.ID) > max {
			max = len(p.ID)
		}
	}

	if err := printfln(s.out, "Processes: "); err != nil {
		return err
	}
	for _, p := range processes {
		err := printfln(s.out, " * %-"+fmt.Sprintf("%d", max+1)+"s %s", p.ID+":", p.Tpl)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) getStdOutErr(id string, verbose bool) (io.Writer, io.Writer) {
	if !verbose {
		return ioutil.Discard, ioutil.Discard
	}

	out := pkgIoutil.NewPrefixWriter(
		s.out,
		fmt.Sprintf("[stdout][%s] ", id),
	)
	err := pkgIoutil.NewPrefixWriter(
		s.err,
		fmt.Sprintf("[stderr][%s] ", id),
	)

	return out, err
}

func (s *Service) filter(r Request) ([]processdef.Process, error) {
	if len(r.Names) > 0 && len(r.Tags) > 0 {
		return nil, fmt.Errorf("cannot combine names and tags")
	}

	inSlice := func(n string, s []string) bool {
		for _, v := range s {
			if v == n {
				return true
			}
		}
		return false
	}

	res := s.processes
	if len(r.Names) > 0 {
		res = processdef.FilterList(res, func(p processdef.Process) bool {
			return inSlice(p.ID, r.Names)
		})
	}
	if len(r.Tags) > 0 {
		res = processdef.FilterList(res, func(p processdef.Process) bool {
			for _, t := range r.Tags {
				if !inSlice(t, p.Tags) {
					return false
				}
			}
			return true
		})
	}

	return res, nil
}

func printfln(w io.Writer, s string, i ...interface{}) error {
	_, err := fmt.Fprintf(w, s+"\n", i...)
	if err != nil {
		return err
	}
	return nil
}
