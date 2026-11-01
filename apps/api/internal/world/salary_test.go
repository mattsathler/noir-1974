package world

import "testing"

func TestDailySalaryCreditsUsesFixedRoleBase(t *testing.T) {
	cases := []struct {
		role string
		want int
	}{
		{role: "accountant", want: 18},
		{role: "MANAGER", want: 26},
		{role: "clerk", want: 11},
		{role: "driver", want: 10},
		{role: "secretary", want: 12},
		{role: "unknown role", want: 0},
	}
	for _, item := range cases {
		t.Run(item.role, func(t *testing.T) {
			if got := DailySalaryCredits(item.role); got != item.want {
				t.Fatalf("DailySalaryCredits(%q) = %d, want %d", item.role, got, item.want)
			}
		})
	}
}
