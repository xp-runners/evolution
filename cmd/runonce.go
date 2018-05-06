package cmd

import "fmt"

type runonce struct {
}

// Runs this commandline
func (r runonce) Run(c *commandline) (int, error) {
	cmd, err := NewCmd(c)
	if err != nil {
		return 255, fmt.Errorf("Could not create process: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return 255, fmt.Errorf("Could not start process: %v", err)
	}

	return exitStatus(cmd.Wait())
}

// Name returns this execution model's name
func (r runonce) Name() string {
	return "default"
}

// String creates a string representation of this execution model
func (r runonce) String() string {
	return "once"
}
