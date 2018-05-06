package cmd

import (
	"bufio"
	"os"
	"strings"
)

type configFile struct {
	Ini    string
	parsed map[string](map[string][]string)
}

func ConfigFile(ini string) *configFile {
	return &configFile{ini, make(map[string](map[string][]string))}
}

func (f configFile) parse() error {
	if len(f.parsed) > 0 {
		return nil
	}

	file, err := os.Open(f.Ini)
	if err != nil {
		return err
	}
	defer file.Close()

	section := "default"
	f.parsed[section] = make(map[string][]string)

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.Trim(line, " \r\n\t")
		if "" == line || ';' == line[0] {
			// Skip
		} else if '[' == line[0] {
			section = line[1 : len(line)-1]
			f.parsed[section] = make(map[string][]string)
		} else {
			pair := strings.SplitN(line, "=", 2)
			f.parsed[section][pair[0]] = append(f.parsed[section][pair[0]], pair[1])
		}
	}

	return nil
}

// Use returns the XP boot path to use
func (f configFile) Use() *string {
	f.parse()

	if prop, ok := f.parsed["default"]["use"]; ok {
		return &(prop[0])
	}
	return nil
}

// Exe returns the PHP executable to fork
func (f configFile) Exe() *string {
	f.parse()

	if rt, ok := f.parsed["default"]["rt"]; ok {
		if prop, ok := f.parsed["runtime@"+rt[0]]["default"]; ok {
			return &(prop[0])
		}
	}
	if prop, ok := f.parsed["runtime"]["default"]; ok {
		return &(prop[0])
	}
	return nil
}

// Extensions returns the PHP extensions to load
func (f configFile) Extensions() []string {
	f.parse()
	r := make([]string, 0)

	if rt, ok := f.parsed["default"]["rt"]; ok {
		if prop, ok := f.parsed["runtime@"+rt[0]]["extension"]; ok {
			r = append(r, prop...)
		}
	}
	if prop, ok := f.parsed["runtime"]["extension"]; ok {
		r = append(r, prop...)
	}
	return r
}

// Extensions returns the PHP extensions to load
func (f configFile) Arguments() map[string]string {
	f.parse()
	r := make(map[string]string)
	merge := func(section map[string][]string) {
		for key, values := range section {
			if key != "default" && key != "extension" {
				r[key] = values[0]
			}
		}
	}

	if rt, ok := f.parsed["default"]["rt"]; ok {
		if section, ok := f.parsed["runtime@"+rt[0]]; ok {
			merge(section)
		}
	}
	if section, ok := f.parsed["runtime"]; ok {
		merge(section)
	}
	return r
}

// String returns a string representation of this config file
func (f configFile) String() string {
	return f.Ini
}
