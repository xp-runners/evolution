package cmd

type run struct {
}

// Arguments formats command arguments for this command
func (r run) Arguments(args []string) []string {
	return args
}

// Modules returns additional modules to load for this command
func (r run) Modules() []string {
	return []string{}
}
