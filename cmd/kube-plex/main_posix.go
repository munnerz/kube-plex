package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

// Catch interrupt and SIGTERM and terminate gracefully
var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

// ProtectSigKill takes measures to shield transcoding process from SIGKILL
func protectSigKill(ctx context.Context) context.Context {
	// Verify that we aren't already running with protections
	protected, _ := strconv.ParseBool(os.Getenv("KUBEPLEX_SIGKILL_PROTECTION"))
	if !protected {
		// Prepare for self exec
		args := os.Args
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Set environment variables for signalling
		cmd.Env = append(os.Environ(),
			"KUBEPLEX_SIGKILL_PROTECTION=true",
			fmt.Sprintf("KUBEPLEX_SIGKILL_PARENT_PID=%d", os.Getpid()),
		)

		// Run command
		err := cmd.Run()

		// If child process returned an error, set error code accordingly
		ecode := 0
		if err != nil {
			if cerr, ok := err.(*exec.ExitError); ok {
				ecode = cerr.ExitCode()
			}
			fmt.Printf("Protected process returned an error: %v\n", err)
		}

		// Quit once subprocess is done
		os.Exit(ecode)
	}

	ppvar := os.Getenv("KUBEPLEX_SIGKILL_PARENT_PID")
	if ppvar == "" {
		fmt.Printf("Parent PID not provided in KUBEPLEX_SIGKILL_PARENT_PID!")
		os.Exit(1)
	}

	ppid, err := strconv.Atoi(ppvar)
	if err != nil {
		fmt.Printf("Invalid parent pid %v: %v", ppvar, err)
		os.Exit(1)
	}

	// FindProcess always returns an os.Process{}
	parent, _ := os.FindProcess(ppid)

	// Set up monitoring for parent process. This is necessary since the parent
	// is killed with SIGKILL and we don't get any signals for this.
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := parent.Signal(syscall.Signal(0))
				if err == os.ErrProcessDone {
					return
				}
			}
		}
	}()

	// Cleanup environment set specifically for the protections
	os.Unsetenv("KUBEPLEX_SIGKILL_PROTECTION")
	os.Unsetenv("KUBEPLEX_SIGKILL_PARENT_PID")

	return ctx
}
