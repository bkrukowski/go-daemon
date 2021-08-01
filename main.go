package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bkrukowski/go-daemon/cmd"
	"github.com/bkrukowski/go-daemon/pkg/cobrautils"
	"github.com/bkrukowski/go-daemon/pkg/lennyface"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	finished := make(chan bool)

	root := cobra.Command{
		Use:           "go-daemon",
		Short:         "",
		Long:          "",
		Version:       fmt.Sprintf("%s %s %s", version, commit, date),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			switch cmd.Name() {
			case "run":

			default:
				return nil
			}

			return cobrautils.NewCancelablePreRun(cancel, finished)(cmd, args)
		},
	}

	root.PersistentFlags().StringP(cobrautils.FlagTimeout, "", "", "timeout, see https://pkg.go.dev/time#ParseDuration")
	root.PersistentFlags().BoolP(cobrautils.FlagVerbose, "v", false, "verbose output")

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
