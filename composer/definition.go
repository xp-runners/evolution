package composer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

var FileName = "composer.json"

type Definition struct {
	Name    string
	Require map[string]string
	Scripts map[string]string `json:"-"`
	Config  struct {
		Vendor string `json:"vendor-dir"`
	}
}

func (c *Definition) VendorDir() string {
	if c.Config.Vendor != "" {
		return c.Config.Vendor
	}
	return VendorDir
}

// File reads a composer.json file.
func File(name string) (*Definition, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	var c struct {
		Definition
		RawScripts map[string]json.RawMessage `json:"scripts"`
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", name, err)
	}
	c.Scripts = make(map[string]string)
	for name, raw := range c.RawScripts {
		var script string
		if json.Unmarshal(raw, &script) == nil && strings.HasPrefix(script, "xp ") {
			c.Scripts[name] = script
		}
	}
	return &c.Definition, nil
}
