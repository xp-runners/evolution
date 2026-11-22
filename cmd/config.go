package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

type iniSection struct {
	keys   []string
	values map[string][]string
}

type iniFile map[string]*iniSection

func readINI(path string) (iniFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	file := make(iniFile)
	section := &iniSection{values: make(map[string][]string)}
	file["default"] = section
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 4096), 1024*1024)
	for n := 1; s.Scan(); n++ {
		line := strings.TrimSpace(strings.TrimPrefix(s.Text(), "\ufeff"))
		switch {
		case line == "", strings.HasPrefix(line, ";"), strings.HasPrefix(line, "#"):
			continue
		case strings.HasPrefix(line, "["):
			if !strings.HasSuffix(line, "]") {
				return nil, fmt.Errorf("%s:%d: invalid section", path, n)
			}
			section = &iniSection{values: make(map[string][]string)}
			file[line[1:len(line)-1]] = section
		default:
			key, value, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key, value = strings.TrimSpace(key), strings.TrimSpace(value)
			if _, exists := section.values[key]; !exists {
				section.keys = append(section.keys, key)
				section.values[key] = nil
			}
			if value == "" {
				continue
			}
			q := value[0]
			if len(value) >= 2 && (q == '"' || q == '\'' || q == '`') && value[len(value)-1] == q {
				value = strings.ReplaceAll(value[1:len(value)-1], "\\"+string(q), string(q))
			}
			section.values[key] = append(section.values[key], value)
		}
	}
	return file, s.Err()
}

func (f iniFile) get(section, key string) (string, bool) {
	if s := f[section]; s != nil && len(s.values[key]) > 0 {
		return s.values[key][0], true
	}
	return "", false
}

type settings struct {
	use        []string
	executable string
	extensions []string
	args       map[string][]string
}

func (c *commandline) settings() (settings, error) {
	s := settings{args: make(map[string][]string)}
	rt, hasRT := os.LookupEnv("XP_RT")
	if use, ok := os.LookupEnv("USE_XP"); ok {
		s.use = translatePaths(".", use, c.home)
	}
	var sources []iniFile
	for _, path := range c.configFiles {
		file, err := readINI(path)
		if errors.Is(err, os.ErrNotExist) && !c.explicitConfig {
			continue
		}
		if err != nil {
			return s, err
		}
		sources = append(sources, file)
		if !hasRT {
			rt, hasRT = file.get("default", "rt")
		}
		if s.use == nil {
			if use, ok := file.get("default", "use"); ok {
				s.use = translatePaths(filepath.Dir(path), use, c.home)
			}
		}
	}

	extensions := make(map[string]struct{})
	for _, file := range sources {
		if s.executable == "" {
			if value, ok := file.get("runtime@"+rt, "default"); ok {
				s.executable = value
			} else if value, ok := file.get("runtime", "default"); ok {
				s.executable = value
			}
		}
		args := make(map[string][]string)
		for _, section := range []string{"runtime", "runtime@" + rt} {
			if data := file[section]; data != nil {
				for key, values := range data.values {
					switch key {
					case "default":
					case "extension":
						for _, ext := range values {
							if _, exists := extensions[ext]; !exists {
								s.extensions = append(s.extensions, ext)
								extensions[ext] = struct{}{}
							}
						}
					default:
						args[key] = values
					}
				}
			}
		}
		for key, values := range args {
			if _, exists := s.args[key]; !exists {
				s.args[key] = values
			}
		}
	}
	if s.executable == "" {
		s.executable = rt
	}
	if s.executable == "" {
		s.executable = "php"
	}
	return s, nil
}

func translatePaths(root, paths, home string) []string {
	result := []string{}
	for _, path := range strings.Split(paths, string(os.PathListSeparator)) {
		if path == "" {
			continue
		}
		if path == "~" || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`) {
			path = filepath.Join(home, path[1:])
		} else if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		if absolute, err := filepath.Abs(path); err == nil {
			path = absolute
		}
		result = append(result, path)
	}
	return result
}

var envReference = regexp.MustCompile(`\\\$|\$\{([^}]+)\}|\$([a-zA-Z0-9_]+)`)

func (c *commandline) expandEnv(env []string) ([]string, error) {
	values := make(map[string]string)
	normalize := func(key string) string {
		if runtime.GOOS == "windows" {
			return strings.ToUpper(key)
		}
		return key
	}
	for _, entry := range env {
		key, value, _ := strings.Cut(entry, "=")
		values[normalize(key)] = value
	}
	files := c.envFiles
	defaults := len(files) == 0
	if defaults {
		files = []string{".env"}
		if name, ok := os.LookupEnv("XP_ENV"); ok {
			files = append(files, ".env."+name)
		}
		files = append(files, ".env.local")
	}
	for _, path := range files {
		file, err := readINI(path)
		if defaults && errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, key := range file["default"].keys {
			raw, _ := file.get("default", key)
			value := envReference.ReplaceAllStringFunc(raw, func(ref string) string {
				if ref == `\$` {
					return "$"
				}
				return values[normalize(strings.Trim(ref[1:], "{}"))]
			})
			values[normalize(key)] = value
			env = append(env, key+"="+value)
		}
	}
	return env, nil
}
