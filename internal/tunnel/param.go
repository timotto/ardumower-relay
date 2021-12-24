package tunnel

import (
	"time"
)

type Parameters struct {
	PingInterval time.Duration `yaml:"ping_interval"`
	PingTimeout  time.Duration `yaml:"ping_timeout"`
	PongTimeout  time.Duration `yaml:"pong_timeout"`
}

func (p *Parameters) Validate() error {
	if p.PingInterval == 0 {
		p.PingInterval = time.Minute
	}

	if p.PingTimeout == 0 {
		p.PingTimeout = 10 * time.Second
	}

	if p.PongTimeout == 0 {
		p.PongTimeout = 10 * time.Second
	}

	return nil
}
