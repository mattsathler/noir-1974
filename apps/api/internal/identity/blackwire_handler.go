package identity

import (
	"errors"
	"log"
	"net/http"
)

func (h *Handler) blackwireProfile(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	reader, supported := h.store.(blackwireStore)
	if !supported {
		h.writeAPIError(w, r, http.StatusServiceUnavailable, "BLACKWIRE_UNAVAILABLE", "The Blackwire account is unavailable.", nil)
		return
	}
	profile, err := reader.blackwireProfile(r.Context(), current.ID, h.worldServerID)
	if errors.Is(err, ErrNoCharacter) {
		h.writeAPIError(w, r, http.StatusConflict, "PLAYER_CHARACTER_REQUIRED", "Create a personal character before opening a Blackwire account.", nil)
		return
	}
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "BLACKWIRE_PROFILE_FAILED", "The Blackwire account could not be loaded.", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"account": profile})
}

func (h *Handler) registerNewBlackwireIdentity(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	registrar, supported := h.store.(blackwireIdentityRegistrar)
	if !supported {
		h.writeAPIError(w, r, http.StatusServiceUnavailable, "BLACKWIRE_UNAVAILABLE", "A new Blackwire identity cannot be registered.", nil)
		return
	}
	profile, err := registrar.registerNewBlackwireIdentity(r.Context(), current.ID, h.worldServerID)
	if errors.Is(err, errBlackwireIdentityNotBanned) {
		h.writeAPIError(w, r, http.StatusConflict, "BLACKWIRE_IDENTITY_ACTIVE", "The current Blackwire identity is still active.", nil)
		return
	}
	if errors.Is(err, ErrNoCharacter) {
		h.writeAPIError(w, r, http.StatusConflict, "PLAYER_CHARACTER_REQUIRED", "Create a personal character before registering a Blackwire identity.", nil)
		return
	}
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "BLACKWIRE_REGISTRATION_FAILED", "A new Blackwire identity could not be registered.", err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"account": profile})
}

func (h *Handler) blackwireMarket(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	reader, supported := h.store.(blackwireStore)
	if !supported {
		h.writeAPIError(w, r, http.StatusServiceUnavailable, "BLACKWIRE_UNAVAILABLE", "The information exchange is unavailable.", nil)
		return
	}
	market, err := reader.blackwireOffers(r.Context(), current.ID, h.worldServerID)
	if errors.Is(err, ErrNoCharacter) {
		h.writeAPIError(w, r, http.StatusConflict, "PLAYER_CHARACTER_REQUIRED", "Create a personal character before opening Blackwire.", nil)
		return
	}
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "BLACKWIRE_MARKET_FAILED", "The information market could not be loaded.", err)
		return
	}
	writeJSON(w, http.StatusOK, market)
}

func (h *Handler) sellBlackwireOffer(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	writer, supported := h.store.(blackwireStore)
	if !supported {
		h.writeAPIError(w, r, http.StatusServiceUnavailable, "BLACKWIRE_UNAVAILABLE", "The information exchange is unavailable.", nil)
		return
	}
	offer, profile, err := writer.sellBlackwireItem(r.Context(), current.ID, h.worldServerID, r.PathValue("boardID"), r.PathValue("itemID"))
	if errors.Is(err, errBlackwireUnavailable) {
		h.writeAPIError(w, r, http.StatusConflict, "BLACKWIRE_OFFER_UNAVAILABLE", "This information is no longer available for sale.", nil)
		return
	}
	if errors.Is(err, errBlackwireBanned) {
		h.writeAPIError(w, r, http.StatusForbidden, "BLACKWIRE_IDENTITY_BANNED", "The Crow has burned this Blackwire identity.", nil)
		return
	}
	if errors.Is(err, ErrNoCharacter) {
		h.writeAPIError(w, r, http.StatusConflict, "PLAYER_ECONOMY_UNAVAILABLE", "The personal account is unavailable.", nil)
		return
	}
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "BLACKWIRE_SALE_FAILED", "The information could not be sold.", err)
		return
	}
	cityID := h.worldFor(r.Context(), current.ID).CityID
	if err := h.recordHackerAction(r.Context(), current.ID, cityID, "blackwire_information_sale", "blackwire:"+offer.ItemID, exposureDeltaForBlackwire(offer.Category)); err != nil {
		log.Printf("record hacker exposure for Blackwire item %s: %v", offer.ItemID, err)
	} else if reader, ok := h.store.(blackwireStore); ok {
		if updated, profileErr := reader.blackwireProfile(r.Context(), current.ID, h.worldServerID); profileErr == nil {
			profile = updated
		}
	}
	if invalidator, ok := h.newspaperReader.(interface{ InvalidateAll() }); ok {
		invalidator.InvalidateAll()
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "sold", "account": profile})
}
