package events

import (
	"errors"
	"time"
)

type TimezoneValidator interface {
	IsValid(timezone string) bool
	GetLocation(timezone string) (*time.Location, error)
}

type timezoneValidator struct{}

func NewTimezoneValidator() TimezoneValidator {
	return &timezoneValidator{}
}

func (v *timezoneValidator) IsValid(timezone string) bool {
	if timezone == "" {
		return false
	}
	_, err := time.LoadLocation(timezone)
	return err == nil
}

func (v *timezoneValidator) GetLocation(timezone string) (*time.Location, error) {
	if timezone == "" {
		return nil, errors.New("timezone cannot be empty")
	}
	return time.LoadLocation(timezone)
}
