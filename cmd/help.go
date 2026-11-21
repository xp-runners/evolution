package cmd

import "strings"

type help struct {
	topic   string
	modules []string
}

// Arguments formats command arguments for this command
func (h help) Arguments(args []string) []string {
	return []string{"xp.runtime.Help", h.topic}
}

// Modules returns additional modules to load for this command
func (h help) Modules() []string {
	return h.modules
}

func resolveHelp(c *commandline) (command, error) {
	resource := func(pkg, topic string) string {
		if topic == "" || topic == "*" {
			topic = "index"
		}
		return "@" + pkg + "/" + topic + ".md"
	}
	if len(c.Args) == 0 {
		return help{topic: resource("xp/runtime", "xp")}, nil
	}
	arg := c.Args[0]
	if strings.HasPrefix(arg, "/") {
		return help{topic: resource("xp/runtime", arg[1:])}, nil
	}
	name, _, hasTopic := strings.Cut(arg, "/")
	if _, ok := commands[name]; ok || name == "list" {
		return help{topic: resource("xp/runtime", arg)}, nil
	}
	p, err := findPlugin(c, name)
	if err != nil {
		return nil, err
	}
	topic := p.Arguments(nil)[0]
	if hasTopic {
		topic = resource("xp/"+p.Package, arg)
	}
	return help{topic: topic, modules: p.Modules()}, nil
}
