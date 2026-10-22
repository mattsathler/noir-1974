package world

import "errors"

var ErrInvalidVisualIdentity = errors.New("invalid modular visual identity")

// VisualIdentity stores stable PNG layer IDs used to assemble a face-only portrait.
// Hair IDs identify already-colored sprites so a ginger identity uses ginger artwork.
type VisualIdentity struct {
	HairID       string   `json:"hairId"`
	SkinColorID  string   `json:"skinColorId"`
	EyeColorID   string   `json:"eyeColorId"`
	FaceDetailID string   `json:"faceDetailId"`
	AccessoryIDs []string `json:"accessoryIds"`
}

func DefaultVisualIdentity() VisualIdentity {
	return VisualIdentity{HairID: "hair_short_brown", SkinColorID: "skin_light", EyeColorID: "eyes_brown", AccessoryIDs: []string{}}
}

var hairLayerIDs = []string{
	"hair_short_black", "hair_short_brown", "hair_short_ginger", "hair_short_blonde", "hair_short_grey",
	"hair_short_alt_black", "hair_short_alt_brown", "hair_short_alt_ginger", "hair_short_alt_blonde", "hair_short_alt_grey",
	"hair_long_black", "hair_long_brown", "hair_long_ginger", "hair_long_blonde", "hair_long_grey",
	"hair_ponytail_black", "hair_ponytail_brown", "hair_ponytail_ginger", "hair_ponytail_blonde", "hair_ponytail_grey",
	"hair_tied_black", "hair_tied_brown", "hair_tied_ginger", "hair_tied_blonde", "hair_tied_grey",
	"hair_tied_long_black", "hair_tied_long_brown", "hair_tied_long_ginger", "hair_tied_long_blonde", "hair_tied_long_grey",
	"hair_messy_black", "hair_messy_brown", "hair_messy_ginger", "hair_messy_blonde", "hair_messy_grey",
	"hair_pigtails_black", "hair_pigtails_brown", "hair_pigtails_ginger", "hair_pigtails_blonde", "hair_pigtails_grey",
	"hair_emo_black", "hair_emo_brown", "hair_emo_ginger", "hair_emo_blonde", "hair_emo_grey",
}
var skinColorLayerIDs = []string{"skin_pale", "skin_light", "skin_medium", "skin_olive", "skin_brown", "skin_dark"}
var eyeColorLayerIDs = []string{"eyes_black", "eyes_blue", "eyes_brown", "eyes_green", "eyes_hazel", "eyes_gray", "eyes_amber"}
var faceDetailLayerIDs = []string{
	"face_beard_black", "face_beard_brown", "face_beard_ginger", "face_beard_blonde", "face_beard_grey",
	"face_beard_goatee_black", "face_beard_goatee_brown", "face_beard_goatee_ginger", "face_beard_goatee_blonde", "face_beard_goatee_grey",
	"face_beard_stubble_black", "face_beard_stubble_brown", "face_beard_stubble_ginger", "face_beard_stubble_blonde", "face_beard_stubble_grey",
	"face_beard_mutton_black", "face_beard_mutton_brown", "face_beard_mutton_ginger", "face_beard_mutton_blonde", "face_beard_mutton_grey",
	"face_moustache_black", "face_moustache_brown", "face_moustache_ginger", "face_moustache_blonde", "face_moustache_grey",
	"face_scar_cheek", "face_scar_brow", "face_stubble", "face_freckles",
}
var accessoryLayerIDs = []string{
	"accessory_glasses_round", "accessory_hat_cloche", "accessory_hat_fedora", "accessory_monocle",
	"accessory_glasses_square", "accessory_hat_bowler",
	"face_glasses_round_black", "face_glasses_round_tortoise", "face_glasses_round_brass", "face_glasses_round_silver", "face_glasses_round_ivory",
	"face_glasses_square_black", "face_glasses_square_tortoise", "face_glasses_square_brass", "face_glasses_square_silver", "face_glasses_square_ivory",
}

func VisualIdentityCatalog() (hair, skinColors, eyeColors, faceDetails, accessories []string) {
	return append([]string(nil), hairLayerIDs...), append([]string(nil), skinColorLayerIDs...), append([]string(nil), eyeColorLayerIDs...), append([]string(nil), faceDetailLayerIDs...), append([]string(nil), accessoryLayerIDs...)
}

func ValidateVisualIdentity(identity VisualIdentity) error {
	if !containsLayer(hairLayerIDs, identity.HairID) || !containsLayer(skinColorLayerIDs, identity.SkinColorID) || !containsLayer(eyeColorLayerIDs, identity.EyeColorID) {
		return ErrInvalidVisualIdentity
	}
	if (identity.FaceDetailID != "" && !containsLayer(faceDetailLayerIDs, identity.FaceDetailID)) || !validateLayerList(identity.AccessoryIDs, accessoryLayerIDs, 1) {
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
