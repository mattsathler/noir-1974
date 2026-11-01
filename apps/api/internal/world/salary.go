package world

import "strings"

var dailyBaseSalaryCredits = map[string]int{
	"accountant":   18,
	"clerk":        11,
	"driver":       10,
	"manager":      26,
	"messenger":    8,
	"receptionist": 9,
	"secretary":    12,
	"typist":       10,
}

// DailySalaryCredits returns the fixed daily base pay for a staff role.
func DailySalaryCredits(role string) int {
	return dailyBaseSalaryCredits[strings.ToLower(strings.TrimSpace(role))]
}
