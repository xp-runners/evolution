package cmd

import "os"

type environment struct {
}

// Environment returns the config instance reading from OS environment
func Environment() *environment {
	return &environment{}
}

// Use returns the XP boot path to use
func (e environment) Use() *string {
	if use, ok := os.LookupEnv("USE_XP"); ok {
		return &use
	}
	return nil
}

// Exe returns the PHP executable to fork
func (e environment) Exe() *string {
	if exe, ok := os.LookupEnv("XP_RT"); ok {
		return &exe
	}
	return nil
}

// Extensions returns the PHP extensions to load
func (e environment) Extensions() []string {
	return []string{}
}

// Extensions returns the PHP extensions to load
func (e environment) Arguments() map[string]string {
	return map[string]string{}
}

// String returns a string representation of environment
func (e environment) String() string {
	return "@env"
}
