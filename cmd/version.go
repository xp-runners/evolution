package cmd

type version struct {
}

// Arguments formats command arguments for this command
func (v version) Arguments(args []string) []string {
	return append([]string{"xp.runtime.Version"}, args...)
}

// Modules returns additional modules to load for this command
func (v version) Modules() []string {
	return []string{}
}
