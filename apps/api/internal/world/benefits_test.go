package world

import (
	"reflect"
	"testing"
)

func TestCompanyBenefitsAreDeterministicAndCompanySpecific(t *testing.T) {
	first := CompanyBenefits("benefits-seed", 7, "company:one")
	second := CompanyBenefits("benefits-seed", 7, "company:one")
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("benefits changed for same seeded company: %+v != %+v", first, second)
	}
	for attempt := 1; attempt < 40; attempt++ {
		other := CompanyBenefits("benefits-seed", 7, "company:"+string(rune('A'+attempt)))
		if !reflect.DeepEqual(first, other) {
			return
		}
	}
	t.Fatal("expected seeded companies to receive different benefit packages")
}
