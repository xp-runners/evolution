package cmd

type help struct {
}

// Arguments formats command arguments for this command
func (h help) Arguments(args []string) []string {
	if len(args) == 0 {
		return []string{"xp.runtime.Help", "@xp/runtime/xp.md"}
	}

	return []string{"xp.runtime.Help", "@xp/runtime/" + args[0] + ".md"}
}

// Modules returns additional modules to load for this command
func (h help) Modules() []string {
	return []string{}
}
