package cmd

import "time"

type runrepeatedly struct {
	Declaration string
	schedule    *schedule
}

func (r runrepeatedly) Run(c *commandline) (int, error) {
	time.Sleep(r.schedule.initial(time.Now()))
	for {
		start := time.Now()
		status, err := (runonce{}).Run(c)
		if err != nil || r.schedule.until(status) {
			return status, err
		}
		time.Sleep(r.schedule.next(start, time.Now()))
	}
}

func (r runrepeatedly) Name() string {
	return "repeat"
}

func (r runrepeatedly) String() string {
	return "repeat(" + r.Declaration + ")"
}
