package cmd

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/bkrukowski/go-daemon/pkg/config/provider"
	"github.com/bkrukowski/go-daemon/pkg/lennyface"
	"github.com/bkrukowski/go-daemon/pkg/process"
	"github.com/bkrukowski/go-daemon/pkg/processdef"
	"github.com/bkrukowski/go-daemon/pkg/runner"
	"github.com/spf13/cobra"
)

//go:embed sample-config.yml
var sampleConfig string

const CleaningUpMsg = "Cleaning up can take up to %s\n"

func NewRun() *cobra.Command {
	const (
		envName         = "GO_DAEMON_CONFIG"
		defaultFileName = "~/.go-daemon.yml"
	)

	var (
		tags              []string
		timeoutFormat     string
		verbose           bool
		ignoreNonZeroCode bool

		long = "Runs processes defined in configuration file.\n" +
			"Default configuration filename is " + defaultFileName + ",\n" +
			"override environment variable " + envName + " to change it.\n" +
			"By default all defined processes will be triggered,\n" +
			"to filter them provide their names as arguments\n" +
			"or use option --tag to filter by tags.\n" +
			"\n" +
			"** Sample config **\n\n" +
			defaultFileName + "\n" +
			"------------------------------------------------------------\n" +
			sampleConfig +
			"------------------------------------------------------------\n"
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "run [process1, process2, ...] [--tag staging, --tag elasticsearch, ...]",
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				if err != nil {
					return
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), lennyface.Sleep)
			}()

			cfg, err := provider.NewDefault(func() (f string) {
				defer func() {
					if !verbose {
						return
					}
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Configuration file: %s\n", f)
				}()

				fn, ok := os.LookupEnv(envName)
				if ok {
					return fn
				}
				return defaultFileName
			}).Provide()

			if err != nil {
				return err
			}

			list := make([]processdef.Process, 0)
			for id, p := range cfg.Processes {
				pd, err := processdef.CreateProcessFromTemplate(id, p.Tags, p.Compiled.Template)
				if err != nil {
					return fmt.Errorf("cannot create process `%s`: %w", id, err)
				}
				list = append(list, pd)
			}
			// cfg.Processes is a map, to achieve predictable results
			// sort by ID always
			sort.Slice(list, func(i, j int) bool {
				return list[i].ID < list[j].ID
			})

			s := runner.NewService(list, cmd.OutOrStdout(), cmd.ErrOrStderr())
			r := runner.Request{
				Names:          args,
				Tags:           tags,
				IgnoreExitCode: ignoreNonZeroCode,
				Verbose:        verbose,
			}
			ctx := cmd.Context()
			if timeoutFormat != "" {
				timeout, err := time.ParseDuration(timeoutFormat)
				if err != nil {
					return fmt.Errorf("invalid timout: %w", err)
				}
				if timeout < 0 {
					return fmt.Errorf("the given timeout is negative: `%s`", timeoutFormat)
				}
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()

				go func() {
					select {
					case <-cmd.Context().Done():
					case <-ctx.Done():
						_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Timed out...")
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), CleaningUpMsg, process.SIGKILLDelay)
					}
				}()
			}
			return s.Run(ctx, r)
		},
	}

	cmd.Flags().StringArrayVarP(&tags, "tag", "t", nil, "filter by tag")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	cmd.Flags().BoolVarP(&ignoreNonZeroCode, "ignore-exit-code", "i", true, "ignore non-zero exit code")
	cmd.Flags().StringVarP(&timeoutFormat, "timeout", "", "", "timeout, see https://pkg.go.dev/time#ParseDuration")

	return cmd
}
