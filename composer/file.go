package composer

import (
	"encoding/json"
	"os"
)

var FileName = "composer.json"

type file struct {
	Name    string
	Require map[string]string
}

// File reads a composer.json file
func File(name string) (*file, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var c file
	dec := json.NewDecoder(f)
	if err := dec.Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
