package world

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
)

// BankAccountNumberForNPC assigns each generated staff member a stable eight-digit account number.
func BankAccountNumberForNPC(seed string, version int, npcID string) string {
	key := strings.TrimPrefix(npcID, "npc:")
	parts := strings.Split(key, "/")
	if seed != "" && version >= GeneratorVersion7 && len(parts) == 10 && parts[2] == "people" && parts[4] == "district" && parts[6] == "company" && parts[8] == "person" {
		cityID := parts[0] + "/" + parts[1]
		city, cityOK := CityByID(cityID)
		companyIndex, companyErr := strconv.Atoi(parts[7])
		slot, slotErr := strconv.Atoi(parts[9])
		cityOrdinal := -1
		for index, candidate := range Cities() {
			if candidate.ID == cityID {
				cityOrdinal = index
				break
			}
		}
		districtOrdinal := -1
		if cityOK {
			citySeed := CitySeed(seed, city.ID)
			for index, district := range districtsForCity(citySeed, version, city) {
				if district.slug == parts[5] {
					districtOrdinal = index
					break
				}
			}
		}
		if cityOrdinal >= 0 && cityOrdinal < 64 && districtOrdinal >= 0 && districtOrdinal < 16 && companyErr == nil && companyIndex > 0 && companyIndex < 64 && slotErr == nil && slot > 0 && slot <= 16 {
			seedDigest := sha256.Sum256([]byte("noir-bank-account/v1/" + seed))
			value := (uint32(seedDigest[0]&0x3f) << 20) | (uint32(cityOrdinal) << 14) | (uint32(districtOrdinal) << 10) | (uint32(companyIndex) << 4) | uint32(slot-1)
			return fmt.Sprintf("%08d", value)
		}
	}
	digest := sha256.Sum256([]byte(fmt.Sprintf("noir-bank-account/v1/%s/%d/%s", seed, version, npcID)))
	value := uint32(digest[0])<<24 | uint32(digest[1])<<16 | uint32(digest[2])<<8 | uint32(digest[3])
	return fmt.Sprintf("%08d", value%100_000_000)
}
