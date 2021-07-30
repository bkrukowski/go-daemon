package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	root := cobra.Command{
		Use:           "go-daemon",
		Short:         "",
		Long:          "",
		Version:       fmt.Sprintf("%s %s %s", version, commit, date),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// see cobra.Command{}.Print
	// it prints to stderr if output is not defined
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)

	root.AddCommand(cmd.NewRun())

	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		v := <-sig
		cancel()
		_, _ = fmt.Fprintln(root.OutOrStdout(), fmt.Sprintf("Received signal \"%s\"...", v))
		_, _ = fmt.Fprintf(root.OutOrStdout(), cmd.CleaningUpMsg, process.SIGKILLDelay)
	}()

	if err := root.ExecuteContext(ctx); err != nil {
		_, _ = fmt.Fprintln(root.ErrOrStderr(), "Error:", err.Error())
		_, _ = fmt.Fprintln(root.ErrOrStderr(), lennyface.Shrug)
		os.Exit(1)
	}
}
