package identity

import (
	"errors"
	"net/http"
)

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
	_, _, err := writer.sellBlackwireItem(r.Context(), current.ID, h.worldServerID, r.PathValue("boardID"), r.PathValue("itemID"))
	if errors.Is(err, errBlackwireUnavailable) {
		h.writeAPIError(w, r, http.StatusConflict, "BLACKWIRE_OFFER_UNAVAILABLE", "This information is no longer available for sale.", nil)
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
	if invalidator, ok := h.newspaperReader.(interface{ InvalidateAll() }); ok {
		invalidator.InvalidateAll()
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "sold"})
}
