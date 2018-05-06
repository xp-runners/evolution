package cmd

import (
	"bufio"
	"fmt"
	"os"
)

type supervise struct {
}

// Runs this commandline
func (r supervise) Run(c *commandline) (int, error) {
	cmd, err := NewCmd(c)
	if err != nil {
		return 255, fmt.Errorf("Could not create process: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return 255, fmt.Errorf("Could not start process: %v", err)
	}

	// TBI: Check whether process dies right after startup
	go func() {
		bufio.NewReader(os.Stdin).ReadString('\n')
		cmd.Process.Kill()
	}()

	return exitStatus(cmd.Wait())
}

// Name returns this execution model's name
func (r supervise) Name() string {
	return "default"
}

// String creates a string representation of this execution model
func (r supervise) String() string {
	return "once"
}
