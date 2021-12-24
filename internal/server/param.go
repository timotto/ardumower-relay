package server

import "fmt"

func (p *Parameters) Validate() error {
	if err := p.Https.validate(); err != nil {
		return fmt.Errorf("invalid https configuration: %w", err)
	}

	if !p.Http.Enabled && !p.Https.Enabled {
		return fmt.Errorf("at least one of http, https must be enabled")
	}

	return nil
}

func (p *HttpsParameters) validate() error {
	if !p.Enabled {
		return nil
	}

	if p.CertFile == "" {
		return fmt.Errorf("cert file cannot be empty")
	}

	if p.KeyFile == "" {
		return fmt.Errorf("key file cannot be empty")
	}

	return nil
}
