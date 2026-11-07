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
	role = strings.ToLower(strings.TrimSpace(role))
	if salary, ok := dailyBaseSalaryCredits[role]; ok {
		return salary
	}
	switch {
	case strings.Contains(role, "director"), strings.Contains(role, "manager"), strings.Contains(role, "partner"), strings.Contains(role, "superintendent"), strings.Contains(role, "master"):
		return 26
	case strings.Contains(role, "engineer"), strings.Contains(role, "physician"), strings.Contains(role, "pilot"), strings.Contains(role, "chemist"), strings.Contains(role, "architect"), strings.Contains(role, "publisher"):
		return 21
	case strings.Contains(role, "foreman"), strings.Contains(role, "supervisor"), strings.Contains(role, "adjuster"), strings.Contains(role, "teacher"), strings.Contains(role, "editor"), strings.Contains(role, "agent"):
		return 16
	case strings.Contains(role, "clerk"), strings.Contains(role, "assistant"), strings.Contains(role, "operator"), strings.Contains(role, "worker"), strings.Contains(role, "cashier"), strings.Contains(role, "driver"), strings.Contains(role, "guard"), strings.Contains(role, "cleaner"), strings.Contains(role, "porter"), strings.Contains(role, "waiter"), strings.Contains(role, "usher"), strings.Contains(role, "miner"), strings.Contains(role, "mechanic"), strings.Contains(role, "technician"), strings.Contains(role, "nurse"), strings.Contains(role, "receptionist"), strings.Contains(role, "secretary"), strings.Contains(role, "typist"), strings.Contains(role, "bookkeeper"), strings.Contains(role, "accountant"):
		return 11
	default:
		return 0
	}
}
