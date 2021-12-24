package config

import "fmt"

type validator interface {
	Validate() error
}

type namedValidator struct {
	N string
	V validator
}

func validate(m []namedValidator) (err error) {
	for _, v := range m {
		if err = validateOne(v.N, v.V); err != nil {
			return
		}
	}

	return nil
}

func validateOne(name string, v validator) error {
	if err := v.Validate(); err != nil {
		return fmt.Errorf("invalid %v configuration: %w", name, err)
	}

	return nil
}
