package daemons

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Daemons struct {
	daemons []Daemon
}

func New(daemons []Daemon) *Daemons {
	return &Daemons{daemons: daemons}
}

type Daemon interface {
	Daemonize(context.Context) error
}

func (d Daemons) Daemonize(ctx context.Context) error {
	g, gctx := errgroup.WithContext(ctx)

	for _, tmp := range d.daemons {
		sd := tmp
		g.Go(func() error {
			return sd.Daemonize(gctx)
		})
	}

	return g.Wait()
}
