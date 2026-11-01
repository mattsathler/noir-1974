package world

import (
	"crypto/sha256"
	"fmt"
)

// CompanyBenefit is an employer policy shown only with the accessed staff directory.
type CompanyBenefit struct {
	Code          string   `json:"code"`
	Name          string   `json:"name"`
	RatePercent   int      `json:"ratePercent"`
	Schedule      string   `json:"schedule"`
	EligibleRoles []string `json:"eligibleRoles"`
}

var seededCompanyBenefits = []CompanyBenefit{
	{Code: "meal_voucher", Name: "Meal voucher", RatePercent: 25, Schedule: "Monday–Friday", EligibleRoles: []string{"accountant", "clerk", "driver", "manager", "messenger", "receptionist", "secretary", "typist"}},
	{Code: "transport_allowance", Name: "Transport allowance", RatePercent: 40, Schedule: "Monday–Friday", EligibleRoles: []string{"accountant", "clerk", "messenger", "receptionist", "secretary", "typist"}},
	{Code: "profit_sharing", Name: "Profit sharing", RatePercent: 200, Schedule: "First Friday of each month", EligibleRoles: []string{"accountant", "manager"}},
}

// BenefitCatalog returns copies of the available policies for payroll generation.
func BenefitCatalog() []CompanyBenefit {
	result := make([]CompanyBenefit, len(seededCompanyBenefits))
	for index, benefit := range seededCompanyBenefits {
		benefit.EligibleRoles = append([]string(nil), benefit.EligibleRoles...)
		result[index] = benefit
	}
	return result
}

// CompanyBenefits returns stable employer policies derived from the world seed and company identity.
func CompanyBenefits(seed string, worldVersion int, companyID string) []CompanyBenefit {
	if seed == "" || worldVersion <= 0 || companyID == "" {
		return []CompanyBenefit{}
	}
	digest := sha256.Sum256([]byte(fmt.Sprintf("noir-company-benefits/v1/%s/%d/%s", seed, worldVersion, companyID)))
	mask := digest[0] & 0b111
	benefits := make([]CompanyBenefit, 0, len(seededCompanyBenefits))
	for index, benefit := range seededCompanyBenefits {
		if mask&(1<<index) == 0 {
			continue
		}
		benefit.EligibleRoles = append([]string(nil), benefit.EligibleRoles...)
		benefits = append(benefits, benefit)
	}
	return benefits
}
