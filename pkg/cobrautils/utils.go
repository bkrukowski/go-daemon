package cobrautils

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const (
	FlagTimeout = "timeout"
	FlagVerbose = "verbose"
)

// NewCancelablePreRun returns func you can assign to cobra.Command{}.PersistentPreRunE or cobra.Command{}.PreRunE.
// It calls cancels func when syscall.SIGINT or syscall.SIGTERM is sent to current process or timeout happens.
//
// Timeout is defined as a string flag with name "timeout'.
// To parse timeout time.ParseDuration will be used.
func NewCancelablePreRun(cancel func(), finished <-chan bool) func(*cobra.Command, []string) error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	return func(cmd *cobra.Command, args []string) (err error) {
		timeoutFormat, err := cmd.Flags().GetString(FlagTimeout)
		if err != nil {
			return err
		}

		verbose, err := cmd.Flags().GetBool(FlagVerbose)
		if err != nil {
			return err
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

			// Do not print message immediately.
			// If finishing all will take less than 1 second, there is no need to print this information.
			t.Reset(time.Second)
			select {
			case <-t.C:
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Cleaning up can take few seconds\n")
			case <-finished:
			}
		}()

		return nil
	}
}
