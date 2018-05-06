package cmd

type runrepeatedly struct {
	Declaration string
}

func (r runrepeatedly) Run(c *commandline) (int, error) {
	return 0, nil
}

func (r runrepeatedly) Name() string {
	return "repeat"
}

func (r runrepeatedly) String() string {
	return "repeat(" + r.Declaration + ")"
}
