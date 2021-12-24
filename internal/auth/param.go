package auth

import "fmt"

type Parameters struct {
	Enabled  bool   `yaml:"enabled"`
	Filename string `yaml:"filename"`
}

func (p *Parameters) Validate() error {
	if p.Enabled && p.Filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	return nil
}
