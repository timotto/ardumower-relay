package auth

import "fmt"

type Parameters struct {
	Enabled    bool   `yaml:"enabled"`
	FreeForAll bool   `yaml:"free_for_all"`
	Filename   string `yaml:"filename"`
}

func (p *Parameters) Validate() error {
	if p.Enabled && p.Filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	if p.Enabled && p.FreeForAll {
		return fmt.Errorf("cannot enable both FreeForAll and static credentials at the same time")
	}
	return nil
}
