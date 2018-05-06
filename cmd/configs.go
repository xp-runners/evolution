package cmd

type configs struct {
	All []config
}

func Configs(all ...config) config {
	return &configs{all}
}

func (c configs) Use() *string {
	for _, config := range c.All {
		if s := config.Use(); s != nil {
			return s
		}
	}
	return nil
}

func (c configs) Exe() *string {
	for _, config := range c.All {
		if s := config.Exe(); s != nil {
			return s
		}
	}
	return nil
}

// Extensions returns the PHP extensions to load
func (c configs) Extensions() []string {
	r := make([]string, 0)
	for _, config := range c.All {
		r = append(r, config.Extensions()...)
	}
	return r
}

// Extensions returns the PHP extensions to load
func (c configs) Arguments() map[string]string {
	r := make(map[string]string)
	for _, config := range c.All {
		for key, value := range config.Arguments() {
			r[key] = value
		}
	}
	return r
}

func (c configs) String() string {
	if len(c.All) == 0 {
		return "all[]"
	}

	s := ""
	for _, config := range c.All {
		s += "," + config.String()
	}
	return "all[" + s[1:] + "]"
}
