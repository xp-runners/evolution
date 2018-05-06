package cmd

type dump struct {
	Mode string
}

// Arguments formats command arguments for this command
func (d dump) Arguments(args []string) []string {
	return append([]string{"xp.runtime.Dump", d.Mode}, args...)
}

// Modules returns additional modules to load for this command
func (d dump) Modules() []string {
	return []string{}
}
