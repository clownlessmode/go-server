package domain

import (
	"regexp"
	"strings"
	"time"
)

var simNumberPattern = regexp.MustCompile(`^\d{10}$`)

type Sim struct {
	Number    string
	Balance   *float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ValidateSimNumber(number string) error {
	number = NormalizeSimNumber(number)
	if !simNumberPattern.MatchString(number) {
		return ErrInvalidSimNumber
	}

	return nil
}

func NormalizeSimNumber(number string) string {
	number = strings.TrimSpace(number)
	number = strings.TrimPrefix(number, "+7")
	number = strings.TrimPrefix(number, "7")
	if len(number) == 11 && strings.HasPrefix(number, "8") {
		number = number[1:]
	}

	return number
}

func NewSim(number string) (Sim, error) {
	number = NormalizeSimNumber(number)
	if err := ValidateSimNumber(number); err != nil {
		return Sim{}, err
	}

	return Sim{Number: number}, nil
}
