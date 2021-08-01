package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bkrukowski/go-daemon/cmd"
	"github.com/bkrukowski/go-daemon/pkg/lennyface"
	"github.com/bkrukowski/go-daemon/pkg/process"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	finished := make(chan bool)

	var (
		timeoutFormat string
		verbose       bool
	)

	root := cobra.Command{
		Use:           "go-daemon",
		Short:         "",
		Long:          "",
		Version:       fmt.Sprintf("%s %s %s", version, commit, date),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if cmd.Name() != "run" {
				return
			}

			d := time.Duration(math.MaxInt64)
			if timeoutFormat != "" {
				d, err = time.ParseDuration(timeoutFormat)
			}
			if err != nil {
				return fmt.Errorf("invalid timeout: %w", err)
			}

			if verbose {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Timeout: %s\n", d)
			}

			go func() {
				t := time.NewTimer(d)
				defer t.Stop()

				select {
				case <-t.C:
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Timed out...\n")
				case v := <-sig:
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Received signal \"%s\"...\n", v)
				}
				cancel()

				t.Reset(time.Second)
				select {
				case <-t.C:
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Cleaning up can take up to %s\n", process.SIGKILLDelay)
				case <-finished:
				}
			}()

			return nil
		},
	}

	root.PersistentFlags().StringVarP(&timeoutFormat, "timeout", "", "", "timeout, see https://pkg.go.dev/time#ParseDuration")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// see cobra.Command{}.Print
	// it prints to stderr if output is not defined
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)

	root.AddCommand(cmd.NewRun())

	err := root.ExecuteContext(ctx)
	close(finished)

	if err != nil {
		_, _ = fmt.Fprintln(root.ErrOrStderr(), "Error:", err.Error())
		_, _ = fmt.Fprintln(root.ErrOrStderr(), lennyface.Shrug)
		os.Exit(1)
	}
}
