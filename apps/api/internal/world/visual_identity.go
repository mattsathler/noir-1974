package world

import "errors"

var ErrInvalidVisualIdentity = errors.New("invalid modular visual identity")

// VisualIdentity stores the six sprite-layer categories selected for a character.
type VisualIdentity struct {
	HairID       string   `json:"hairId"`
	SkinID       string   `json:"skinId"`
	EyesID       string   `json:"eyesId"`
	FaceID       string   `json:"faceId"`
	BeardID      string   `json:"beardId"`
	AccessoryIDs []string `json:"accessoryIds"`
}

type VisualIdentityOptions struct {
	Accessories []string `json:"accessories"`
	Beard       []string `json:"beard"`
	Eyes        []string `json:"eyes"`
	Face        []string `json:"face"`
	Hair        []string `json:"hair"`
	Skin        []string `json:"skin"`
}

func DefaultVisualIdentity() VisualIdentity {
	return VisualIdentity{HairID: "hair_short_brown", SkinID: "skin_white", EyesID: "eyes_brown", AccessoryIDs: []string{}}
}

var hairLayerIDs = []string{
	"hair_short_black", "hair_short_brown", "hair_short_ginger", "hair_short_blonde", "hair_short_grey",
	"hair_long_black", "hair_long_brown", "hair_long_ginger", "hair_long_blonde", "hair_long_grey",
	"hair_long_wavy_black", "hair_long_wavy_brown", "hair_long_wavy_ginger", "hair_long_wavy_blonde", "hair_long_wavy_grey",
	"hair_long_curly_black", "hair_long_curly_brown", "hair_long_curly_ginger", "hair_long_curly_blonde", "hair_long_curly_grey",
	"hair_messy_black", "hair_messy_brown", "hair_messy_ginger", "hair_messy_blonde", "hair_messy_grey",
	"hair_wavy_black", "hair_wavy_brown", "hair_wavy_ginger", "hair_wavy_blonde", "hair_wavy_grey",
	"hair_bob_black", "hair_bob_brown", "hair_bob_ginger", "hair_bob_blonde", "hair_bob_grey",
	"hair_slicked_black", "hair_slicked_brown", "hair_slicked_ginger", "hair_slicked_blonde", "hair_slicked_grey",
	"hair_mohawk_black", "hair_mohawk_brown", "hair_mohawk_ginger", "hair_mohawk_blonde", "hair_mohawk_grey",
	"hair_balding_black", "hair_balding_brown", "hair_balding_ginger", "hair_balding_blonde", "hair_balding_grey",
	"hair_sparse_sides_black", "hair_sparse_sides_brown", "hair_sparse_sides_ginger", "hair_sparse_sides_blonde", "hair_sparse_sides_grey",
}
var skinLayerIDs = []string{"skin_white", "skin_pale", "skin_light", "skin_medium", "skin_olive", "skin_brown", "skin_dark"}
var eyeLayerIDs = []string{"eyes_black", "eyes_blue", "eyes_brown", "eyes_green", "eyes_hazel", "eyes_gray", "eyes_amber"}
var faceLayerIDs = []string{"face_scar_short"}
var beardLayerIDs = []string{
	"beard_moustache_black", "beard_moustache_brown", "beard_moustache_ginger", "beard_moustache_blonde", "beard_moustache_grey",
	"beard_mutton_black", "beard_mutton_brown", "beard_mutton_ginger", "beard_mutton_blonde", "beard_mutton_grey",
	"beard_soul_patch_black", "beard_soul_patch_brown", "beard_soul_patch_ginger", "beard_soul_patch_blonde", "beard_soul_patch_grey",
	"beard_goatee_black", "beard_goatee_brown", "beard_goatee_ginger", "beard_goatee_blonde", "beard_goatee_grey",
	"beard_full_black", "beard_full_brown", "beard_full_ginger", "beard_full_blonde", "beard_full_grey",
}
var accessoryLayerIDs = []string{
	"accessories_glasses_round_black", "accessories_glasses_round_tortoise", "accessories_glasses_round_brass", "accessories_glasses_round_silver", "accessories_glasses_round_ivory",
	"accessories_glasses_squared_black", "accessories_glasses_squared_tortoise", "accessories_glasses_squared_brass", "accessories_glasses_squared_silver", "accessories_glasses_squared_ivory",
	"accessories_monocle_black", "accessories_monocle_tortoise", "accessories_monocle_brass", "accessories_monocle_silver", "accessories_monocle_ivory",
}

func VisualIdentityCatalog() VisualIdentityOptions {
	return VisualIdentityOptions{
		Accessories: append([]string(nil), accessoryLayerIDs...),
		Beard:       append([]string(nil), beardLayerIDs...),
		Eyes:        append([]string(nil), eyeLayerIDs...),
		Face:        append([]string(nil), faceLayerIDs...),
		Hair:        append([]string(nil), hairLayerIDs...),
		Skin:        append([]string(nil), skinLayerIDs...),
	}
}

func ValidateVisualIdentity(identity VisualIdentity) error {
	if !containsLayer(hairLayerIDs, identity.HairID) || !containsLayer(skinLayerIDs, identity.SkinID) || !containsLayer(eyeLayerIDs, identity.EyesID) {
		return ErrInvalidVisualIdentity
	}
	if (identity.FaceID != "" && !containsLayer(faceLayerIDs, identity.FaceID)) ||
		(identity.BeardID != "" && !containsLayer(beardLayerIDs, identity.BeardID)) ||
		!validateLayerList(identity.AccessoryIDs, accessoryLayerIDs, 1) {
		return ErrInvalidVisualIdentity
	}
	return nil
}

func validateLayerList(selected, allowed []string, limit int) bool {
	if len(selected) > limit {
		return false
	}
	seen := make(map[string]struct{}, len(selected))
	for _, id := range selected {
		if !containsLayer(allowed, id) {
			return false
		}
		if _, ok := seen[id]; ok {
			return false
		}
		seen[id] = struct{}{}
	}
	return true
}

func containsLayer(allowed []string, selected string) bool {
	for _, id := range allowed {
		if id == selected {
			return true
		}
	}
	return false
}
