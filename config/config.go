package config

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Args          []string
	BranchIgnores []string `json:"branch_ignores"`
	Command       string
}

func (c *Config) IsValid(line string) bool {

	for _, s := range c.BranchIgnores {
		if strings.Contains(line, s) {
			return false
		}
	}

	return true
}

func Initialize(args []string) (conf *Config, err error) {

	if len(args) == 0 {
		return nil, errors.New("Command is required.")
	}

	cmd := args[0]

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(home, ".config", "pecogit")
	if info, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, fs.ModePerm); err != nil {
				return nil, err
			}
		} else if !info.IsDir() {
			return nil, err
		}
	}

	path := filepath.Join(dir, "config.json")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := writeInitialTemplate(path); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	bytes, err := io.ReadAll(fp)
	if err != nil {
		return nil, err
	}

	conf = &Config{
		Args:    args,
		Command: cmd,
	}

	json.Unmarshal(bytes, conf)

	return conf, nil
}

func writeInitialTemplate(path string) error {
	s := `{
    "branch_ignores":[]
}`
	err := os.WriteFile(path, []byte(s), os.ModePerm)
	return err
}
