package cmd

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/bkrukowski/go-daemon/pkg/cobrautils"
	"github.com/bkrukowski/go-daemon/pkg/config/provider"
	"github.com/bkrukowski/go-daemon/pkg/lennyface"
	"github.com/bkrukowski/go-daemon/pkg/processdef"
	"github.com/bkrukowski/go-daemon/pkg/runner"
	"github.com/spf13/cobra"
)

//go:embed help.txt
var help string

func mustPrintf(w io.Writer, s string, i ...interface{}) {
	_, err := fmt.Fprintf(w, s, i...)
	if err != nil {
		panic(err)
	}
}

func NewRun() *cobra.Command {
	const (
		envName         = "GO_DAEMON_CONFIG"
		defaultFileName = "~/.go-daemon.yml"
	)

	var (
		tags              []string
		ignoreNonZeroCode bool
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "run [process1, process2, ...] [--tag staging, --tag elasticsearch, ...]",
		Long:  help,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			verbose, err := cmd.Flags().GetBool(cobrautils.FlagVerbose)
			if err != nil {
				return
			}

			defer func() {
				if err != nil {
					return
				}
				mustPrintf(cmd.OutOrStdout(), lennyface.Sleep+"\n")
			}()

			finished := make(chan bool)
			defer close(finished)

			cfg, err := provider.NewDefault(func() (f string) {
				defer func() {
					if !verbose {
						return
					}
					mustPrintf(cmd.OutOrStdout(), "Configuration file: %s\n", f)
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

			return s.Run(cmd.Context(), r)
		},
	}

	cmd.Flags().StringArrayVarP(&tags, "tag", "t", nil, "filter by tag")
	cmd.Flags().BoolVarP(&ignoreNonZeroCode, "ignore-exit-code", "i", true, "ignore non-zero exit code")

	return cmd
}
