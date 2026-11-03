package identity

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var blackwireUsernamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{2,23}$`)

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
		h.writeAPIError(w, r, http.StatusConflict, "PLAYER_PROFILE_REQUIRED", "Complete your Meridian profile before opening a Blackwire account.", nil)
		return
	}
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "BLACKWIRE_PROFILE_FAILED", "The Blackwire account could not be loaded.", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"account": profile})
}

func (h *Handler) registerBlackwireProfile(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	var request struct {
		Username string `json:"username"`
		AvatarID string `json:"avatarId"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		h.writeAPIError(w, r, http.StatusBadRequest, "BLACKWIRE_PROFILE_INVALID", "Provide a valid username and portrait.", err)
		return
	}
	request.Username = strings.TrimSpace(request.Username)
	if !blackwireUsernamePattern.MatchString(request.Username) {
		h.writeAPIError(w, r, http.StatusBadRequest, "BLACKWIRE_USERNAME_INVALID", "Use 3–24 letters, numbers, dots, underscores or hyphens; begin with a letter or number.", nil)
		return
	}
	writer, supported := h.store.(blackwireProfileWriter)
	if !supported {
		h.writeAPIError(w, r, http.StatusServiceUnavailable, "BLACKWIRE_UNAVAILABLE", "A Blackwire profile cannot be registered.", nil)
		return
	}
	profile, err := writer.registerBlackwireProfile(r.Context(), current.ID, h.worldServerID, request.Username, request.AvatarID)
	switch {
	case errors.Is(err, errBlackwireAvatarUnavailable):
		h.writeAPIError(w, r, http.StatusBadRequest, "BLACKWIRE_PORTRAIT_UNAVAILABLE", "Select a portrait from the published catalogue.", err)
	case errors.Is(err, errBlackwireUsernameTaken):
		h.writeAPIError(w, r, http.StatusConflict, "BLACKWIRE_USERNAME_TAKEN", "That username is already in use on this server.", err)
	case errors.Is(err, errBlackwireProfileActive):
		h.writeAPIError(w, r, http.StatusConflict, "BLACKWIRE_PROFILE_ACTIVE", "This Blackwire profile is already registered.", err)
	case errors.Is(err, errBlackwireBanned):
		h.writeAPIError(w, r, http.StatusForbidden, "BLACKWIRE_IDENTITY_BANNED", "The Crow has burned this Blackwire identity.", err)
	case errors.Is(err, ErrNoCharacter):
		h.writeAPIError(w, r, http.StatusConflict, "PLAYER_PROFILE_REQUIRED", "Complete your Meridian profile before opening a Blackwire account.", err)
	case err != nil:
		h.writeAPIError(w, r, http.StatusInternalServerError, "BLACKWIRE_PROFILE_FAILED", "The Blackwire profile could not be registered.", err)
	default:
		writeJSON(w, http.StatusCreated, map[string]any{"account": profile})
	}
}

func (h *Handler) fetchBlackwireApp(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	store, supported := h.store.(blackwireApplicationStore)
	if !supported {
		h.writeAPIError(w, r, http.StatusServiceUnavailable, "BLACKWIRE_DOWNLOADS_UNAVAILABLE", "The software catalogue is unavailable.", nil)
		return
	}
	profile, err := store.fetchBlackwireApp(r.Context(), current.ID, h.worldServerID, r.PathValue("appID"))
	switch {
	case errors.Is(err, errBlackwirePackageUnavailable):
		h.writeAPIError(w, r, http.StatusNotFound, "BLACKWIRE_PACKAGE_UNAVAILABLE", "This package is not published by Blackwire.", err)
	case errors.Is(err, errBlackwireProfileRequired):
		h.writeAPIError(w, r, http.StatusConflict, "BLACKWIRE_PROFILE_REQUIRED", "Register a Blackwire profile before downloading software.", err)
	case errors.Is(err, errBlackwireBanned):
		h.writeAPIError(w, r, http.StatusForbidden, "BLACKWIRE_IDENTITY_BANNED", "The Crow has burned this Blackwire identity.", err)
	case errors.Is(err, ErrNoCharacter):
		h.writeAPIError(w, r, http.StatusConflict, "PLAYER_PROFILE_REQUIRED", "Complete your Meridian profile before opening a Blackwire account.", err)
	case err != nil:
		h.writeAPIError(w, r, http.StatusInternalServerError, "BLACKWIRE_FETCH_FAILED", "The package could not be fetched.", err)
	default:
		writeJSON(w, http.StatusOK, map[string]any{"status": "fetched", "appId": r.PathValue("appID"), "account": profile})
	}
}

func (h *Handler) installBlackwireApp(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	store, supported := h.store.(blackwireApplicationStore)
	if !supported {
		h.writeAPIError(w, r, http.StatusServiceUnavailable, "BLACKWIRE_INSTALL_UNAVAILABLE", "The software installer is unavailable.", nil)
		return
	}
	profile, err := store.installBlackwireApp(r.Context(), current.ID, h.worldServerID, r.PathValue("appID"))
	switch {
	case errors.Is(err, errBlackwirePackageUnavailable):
		h.writeAPIError(w, r, http.StatusNotFound, "BLACKWIRE_PACKAGE_UNAVAILABLE", "This package is not published by Blackwire.", err)
	case errors.Is(err, errBlackwirePackageNotFetched):
		h.writeAPIError(w, r, http.StatusConflict, "BLACKWIRE_PACKAGE_NOT_FETCHED", "Fetch this package before installing it.", err)
	case errors.Is(err, errBlackwireProfileRequired):
		h.writeAPIError(w, r, http.StatusConflict, "BLACKWIRE_PROFILE_REQUIRED", "Register a Blackwire profile before installing software.", err)
	case errors.Is(err, errBlackwireBanned):
		h.writeAPIError(w, r, http.StatusForbidden, "BLACKWIRE_IDENTITY_BANNED", "The Crow has burned this Blackwire identity.", err)
	case errors.Is(err, ErrNoCharacter):
		h.writeAPIError(w, r, http.StatusConflict, "PLAYER_PROFILE_REQUIRED", "Complete your Meridian profile before opening a Blackwire account.", err)
	case err != nil:
		h.writeAPIError(w, r, http.StatusInternalServerError, "BLACKWIRE_INSTALL_FAILED", "The package could not be installed.", err)
	default:
		writeJSON(w, http.StatusOK, map[string]any{"status": "installed", "account": profile})
	}
}

func (h *Handler) uninstallBlackwireApp(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	store, supported := h.store.(blackwireApplicationUninstaller)
	if !supported {
		h.writeAPIError(w, r, http.StatusServiceUnavailable, "BLACKWIRE_UNINSTALL_UNAVAILABLE", "The software uninstaller is unavailable.", nil)
		return
	}
	appID := r.PathValue("appID")
	profile, err := store.uninstallBlackwireApp(r.Context(), current.ID, h.worldServerID, appID)
	switch {
	case errors.Is(err, errBlackwirePackageUnavailable):
		h.writeAPIError(w, r, http.StatusNotFound, "BLACKWIRE_PACKAGE_UNAVAILABLE", "This application is not available for removal.", err)
	case errors.Is(err, errBlackwireAppPreinstalled):
		h.writeAPIError(w, r, http.StatusConflict, "BLACKWIRE_APP_PREINSTALLED", "Applications included with the computer cannot be removed.", err)
	case errors.Is(err, errBlackwireAppNotInstalled):
		h.writeAPIError(w, r, http.StatusConflict, "BLACKWIRE_APP_NOT_INSTALLED", "This application is not installed.", err)
	case errors.Is(err, errBlackwireProfileRequired):
		h.writeAPIError(w, r, http.StatusConflict, "BLACKWIRE_PROFILE_REQUIRED", "Register a Blackwire profile before managing software.", err)
	case errors.Is(err, errBlackwireBanned):
		h.writeAPIError(w, r, http.StatusForbidden, "BLACKWIRE_IDENTITY_BANNED", "The Crow has burned this Blackwire identity.", err)
	case errors.Is(err, ErrNoCharacter):
		h.writeAPIError(w, r, http.StatusConflict, "PLAYER_PROFILE_REQUIRED", "Complete your Meridian profile before managing software.", err)
	case err != nil:
		h.writeAPIError(w, r, http.StatusInternalServerError, "BLACKWIRE_UNINSTALL_FAILED", "The application could not be removed.", err)
	default:
		writeJSON(w, http.StatusOK, map[string]any{"status": "uninstalled", "appId": appID, "account": profile})
	}
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
		h.writeAPIError(w, r, http.StatusConflict, "PLAYER_PROFILE_REQUIRED", "Complete your Meridian profile before registering a Blackwire identity.", nil)
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
		h.writeAPIError(w, r, http.StatusConflict, "PLAYER_PROFILE_REQUIRED", "Complete your Meridian profile before opening Blackwire.", nil)
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
