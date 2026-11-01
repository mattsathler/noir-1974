package world

import "testing"

func TestBankAccountNumbersAreEightDigitsStableAndUniqueForSeededStaff(t *testing.T) {
	seed := "bank-account-seed"
	base, err := Generate(seed, GeneratorVersion8)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	seen := map[string]string{}
	for _, relation := range base.EmploymentRelationships {
		account := BankAccountNumberForNPC(seed, GeneratorVersion8, relation.EmployeeNPCID)
		if len(account) != 8 {
			t.Fatalf("account number %q does not contain eight digits", account)
		}
		for _, digit := range account {
			if digit < '0' || digit > '9' {
				t.Fatalf("account number %q contains a non-digit", account)
			}
		}
		if existing, duplicate := seen[account]; duplicate && existing != relation.EmployeeNPCID {
			t.Fatalf("NPCs %q and %q share account %s", existing, relation.EmployeeNPCID, account)
		}
		seen[account] = relation.EmployeeNPCID
		if repeated := BankAccountNumberForNPC(seed, GeneratorVersion8, relation.EmployeeNPCID); repeated != account {
			t.Fatalf("account number changed for %q: %s != %s", relation.EmployeeNPCID, repeated, account)
		}
	}
	if len(seen) < 10 {
		t.Fatalf("only generated %d distinct employee accounts", len(seen))
	}
}
