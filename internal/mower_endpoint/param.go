package mower_endpoint

import "fmt"

func (p *Parameters) Validate() error {
	if p.ReadBufferSize == 0 {
		p.ReadBufferSize = 1024
	}

	if p.WriteBufferSize == 0 {
		p.WriteBufferSize = 1024
	}

	if err := p.Tunnel.Validate(); err != nil {
		return fmt.Errorf("invalid tunnel configuration: %w", err)
	}

	return nil
}
