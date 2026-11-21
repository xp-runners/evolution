package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type schedule struct {
	mode     string
	interval time.Duration
	times    []time.Duration
	until    func(int) bool
}

func parseSchedule(spec string) (*schedule, error) {
	s := &schedule{until: func(int) bool { return false }}
	for _, part := range strings.Split(spec, ",") {
		f := strings.Fields(part)
		if len(f) == 0 {
			continue
		}
		bad := func() (*schedule, error) { return nil, fmt.Errorf("repeat: invalid definition %q", part) }
		switch f[0] {
		case "forever":
			s.until = func(int) bool { return false }
		case "immediately":
			s.interval = 0
		case "after", "every":
			if len(f) != 2 {
				return bad()
			}
			d, ok := parseSpan(f[1], false)
			if !ok {
				return bad()
			}
			s.mode, s.interval = f[0], d
		case "at":
			if len(f) < 2 {
				return bad()
			}
			s.mode, s.times = "at", nil
			for _, arg := range f[1:] {
				d, ok := parseSpan(arg, true)
				if !ok {
					return bad()
				}
				s.times = append(s.times, d)
			}
			sort.Slice(s.times, func(i, j int) bool { return s.times[i] < s.times[j] })
		case "until":
			if len(f) != 2 {
				return bad()
			}
			switch f[1] {
			case "success":
				s.until = func(code int) bool { return code == 0 }
			case "error":
				s.until = func(code int) bool { return code != 0 }
			default:
				codes := make(map[int]bool)
				for _, value := range strings.Split(f[1], "|") {
					code, err := strconv.Atoi(value)
					if err != nil {
						return bad()
					}
					codes[code] = true
				}
				s.until = func(code int) bool { return codes[code] }
			}
		default:
			return bad()
		}
	}
	return s, nil
}

func parseSpan(input string, clock bool) (time.Duration, bool) {
	parts := strings.Split(input, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, false
	}
	var n [3]int64
	offset := 0
	if len(parts) == 2 && !clock {
		offset = 1
	}
	for i, part := range parts {
		value, err := strconv.ParseUint(part, 10, 31)
		if err != nil {
			return 0, false
		}
		n[i+offset] = int64(value)
	}
	if clock && (n[0] >= 24 || n[1] >= 60 || n[2] >= 60) {
		return 0, false
	}
	seconds := n[0]*3600 + n[1]*60 + n[2]
	if seconds > (1<<63-1)/int64(time.Second) {
		return 0, false
	}
	return time.Duration(seconds) * time.Second, true
}

func (s schedule) initial(now time.Time) time.Duration {
	if s.mode == "at" {
		return s.target(now, true).Sub(now)
	}
	return 0
}

func (s schedule) next(start, end time.Time) time.Duration {
	switch s.mode {
	case "after":
		return s.interval
	case "every":
		return s.interval - end.Sub(start)
	case "at":
		return s.target(start, false).Sub(end)
	default:
		return 0
	}
}

func (s schedule) target(now time.Time, inclusive bool) time.Time {
	makeTime := func(day time.Time, span time.Duration) time.Time {
		return time.Date(day.Year(), day.Month(), day.Day(), int(span/time.Hour),
			int(span/time.Minute)%60, int(span/time.Second)%60, 0, day.Location())
	}
	for _, span := range s.times {
		target := makeTime(now, span)
		if target.After(now) || inclusive && target.Equal(now) {
			return target
		}
	}
	return makeTime(now.AddDate(0, 0, 1), s.times[0])
}
