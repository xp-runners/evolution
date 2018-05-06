package cmd

type eval struct {
}

func (e eval) Arguments(args []string) []string {
	return append([]string{"xp.runtime.Evaluate"}, args...)
}

// Modules returns additional modules to load for this command
func (e eval) Modules() []string {
	return []string{}
}
