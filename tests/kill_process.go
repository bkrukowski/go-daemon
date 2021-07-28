package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const (
	timeout = time.Second * 3
)

var (
	devNull *os.File
	//cmdName = "bash"
	//cmdArgs = []string{"-c", "sleep 1812 || echo 1"}
	cmdName = "echo"
	cmdArgs = []string{"5"}
)

func init() {
	var err error
	devNull, err = os.OpenFile("/dev/null", os.O_APPEND, os.ModeAppend)
	if err != nil {
		panic(fmt.Sprintf("Could not open /dev/null: %s", err.Error()))
	}
}

func processExists() bool {
	cmd := exec.Command("bash", "-c", "ps aux | grep sleep | grep 1812 | grep -v grep")
	mustStartCmd(cmd)
	err := cmd.Wait()
	if err == nil { // exit code == 0
		return true
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() == 1 {
			return false
		}
		panic(fmt.Sprintf("Unexpected status code: %d", exitErr.ExitCode()))
	}
	panic(fmt.Sprintf("Unexpected error: %s", err.Error()))
}

func mustStartCmd(cmd *exec.Cmd) {
	if err := cmd.Start(); err != nil {
		panic(fmt.Sprintf("Could not start process: %s", err.Error()))
	}
}

func killByContext() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		fmt.Printf("Could not start process: %s\n", err)
	}

	<-ctx.Done()
	fmt.Println("Done")
}

func killByProcessKill() {
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = devNull
	mustStartCmd(cmd)
	time.Sleep(time.Second)
	fmt.Println("Killing process...")
	err := cmd.Process.Kill()
	if err != nil {
		fmt.Printf("Err: %s\n", err.Error())
		return
	}
}

func killBySyscallKill() {
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = devNull
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	mustStartCmd(cmd)
	time.Sleep(time.Second)
	fmt.Println("Killing process...")

	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		panic(fmt.Sprintf("syscall.Getpgid(): %s", err.Error()))
	}
	if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
		panic(fmt.Sprintf("syscall.Kill(): %s", err.Error()))
	}
	if err := cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			panic(fmt.Sprintf("cmd.Wait(): %s", err.Error()))
		}
	}
}

func main() {
	if processExists() {
		fmt.Println("Process exists")
		fmt.Println("Execute 'pkill -f \"sleep 1812\"' and try again")
		os.Exit(1)
	}

	killBy := "context"
	if len(os.Args) >= 2 {
		killBy = os.Args[1]
	}
	fmt.Printf("Creating process `%s %s`\n", cmdName, strings.Join(cmdArgs, " "))
	switch killBy {
	case "context":
		fmt.Println("Kill by context")
		killByContext()
	case "processKill":
		fmt.Println("Kill by Process{}.Kill()")
		killByProcessKill()
	case "syscallKill":
		fmt.Println("Kill by syscall.Kill()")
		killBySyscallKill()
	default:
		fmt.Println("Invalid argument, provide `context`, `processKill` or `syscallKill`")
		os.Exit(1)
	}

	if processExists() {
		fmt.Println("Process exists")
		os.Exit(1)
	}

	fmt.Println("Process has been successfully killed")
}
