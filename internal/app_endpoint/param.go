package app_endpoint

import "time"

type Parameters struct {
	Timeout time.Duration `yaml:"timeout"`
}

func (p *Parameters) Validate() error {
	if p.Timeout == 0 {
		p.Timeout = 10 * time.Second
	}

	return nil
}
