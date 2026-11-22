package cmd

type entry struct {
  Class string
}

// Arguments formats command arguments for this command
func (e entry) Arguments(args []string) []string {
 return append([]string{e.Class}, args...) 
}

// Modules returns additional modules to load for this command
func (e entry) Modules() []string {
  return nil
}
