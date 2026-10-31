package identity

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"noir-1974/apps/api/internal/world"
)

type stockPortfolioReader interface {
	playerStockPortfolio(context.Context, string, string) (playerProgression, []stockPosition, error)
}

type stockOrderWriter interface {
	executeStockOrder(context.Context, string, string, stockListing, stockOrderInput) (int64, error)
}

func (h *Handler) playerProgression(ctx context.Context, accountID string) *playerProgression {
	reader, ok := h.store.(stockPortfolioReader)
	if !ok || h.worldServerID == "" {
		return nil
	}
	progression, _, err := reader.playerStockPortfolio(ctx, accountID, h.worldServerID)
	if errors.Is(err, ErrNoCharacter) || err != nil {
		return nil
	}
	return &progression
}

func (h *Handler) stockMarket(w http.ResponseWriter, r *http.Request) {
	current, ok := h.authorizedPlayer(w, r)
	if !ok {
		return
	}
	reader := h.worldReader.(companySiteIndexReader)
	city, exists := world.CityByID(h.worldFor(r.Context(), current.ID).CityID)
	if !exists {
		h.writeAPIError(w, r, http.StatusConflict, "PROXY_JURISDICTION_UNAVAILABLE", "Connect a Proxy in a country before opening its stock exchange.", nil)
		return
	}
	page := 1
	if value := strings.TrimSpace(r.URL.Query().Get("page")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			h.writeAPIError(w, r, http.StatusBadRequest, "INVALID_MARKET_PAGE", "Market page must be a positive number.", nil)
			return
		}
		page = parsed
	}
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	companies, err := reader.CompanySitesInCountry(r.Context(), h.worldServerID, city.CountryID)
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "MARKET_LIST_FAILED", "National company listings could not be loaded.", err)
		return
	}
	if query != "" {
		filtered := companies[:0]
		for _, record := range companies {
			if strings.Contains(strings.ToLower(record.Company.Name), query) || strings.Contains(strings.ToLower(stockSymbol(record.Company.Name)), query) {
				filtered = append(filtered, record)
			}
		}
		companies = filtered
		page = 1
	}
	start := len(companies)
	if page <= len(companies)/stockPageSize+1 {
		start = (page - 1) * stockPageSize
	}
	end := min(start+stockPageSize, len(companies))
	listings := make([]stockListing, 0, end-start)
	for _, record := range companies[start:end] {
		listings = append(listings, stockListing{
			CompanyID: record.Company.ID, CompanyName: record.Company.Name,
			CountryID: city.CountryID, CityName: record.Address.City,
			PriceCents: stockQuoteCents(city.CountryID, record.Company.ID),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"countryId": city.CountryID, "countryName": city.Country,
		"page": page, "pageSize": stockPageSize, "total": len(companies),
		"listings": listings,
	})
}

func stockSymbol(name string) string {
	words := strings.FieldsFunc(strings.ToUpper(name), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) })
	letters := make([]rune, 0, 3)
	suffix := ""
	for _, word := range words {
		numeric := true
		for _, r := range word {
			if !unicode.IsNumber(r) {
				numeric = false
				break
			}
		}
		if numeric {
			suffix = word
			continue
		}
		if len(letters) < 3 {
			for _, r := range word {
				letters = append(letters, r)
				break
			}
		}
	}
	if len(letters) == 0 {
		letters = []rune{'C', 'O'}
	}
	return string(letters) + suffix
}

func (h *Handler) stockPortfolio(w http.ResponseWriter, r *http.Request) {
	current, ok := h.authorizedPlayer(w, r)
	if !ok {
		return
	}
	reader, ok := h.store.(stockPortfolioReader)
	if !ok {
		h.writeAPIError(w, r, http.StatusServiceUnavailable, "STOCK_PORTFOLIO_UNAVAILABLE", "Player portfolio services are temporarily unavailable.", nil)
		return
	}
	progression, positions, err := reader.playerStockPortfolio(r.Context(), current.ID, h.worldServerID)
	if errors.Is(err, ErrNoCharacter) {
		h.writeAPIError(w, r, http.StatusConflict, "PLAYER_CHARACTER_REQUIRED", "Create a character before using the stock exchange.", nil)
		return
	}
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "STOCK_PORTFOLIO_FAILED", "Your stock portfolio could not be loaded.", err)
		return
	}
	if index, available := h.worldReader.(companySiteIndexReader); available {
		countries := make(map[string]struct{})
		for _, position := range positions {
			countries[position.CountryID] = struct{}{}
		}
		companyNames := make(map[string]string, len(positions))
		for countryID := range countries {
			companies, loadErr := index.CompanySitesInCountry(r.Context(), h.worldServerID, countryID)
			if loadErr != nil {
				h.writeAPIError(w, r, http.StatusInternalServerError, "STOCK_PORTFOLIO_FAILED", "Company names for your portfolio could not be loaded.", loadErr)
				return
			}
			for _, company := range companies {
				companyNames[company.Company.ID] = company.Company.Name
			}
		}
		for positionIndex := range positions {
			if name := companyNames[positions[positionIndex].CompanyID]; name != "" {
				positions[positionIndex].CompanyName = name
			}
		}
		sort.Slice(positions, func(i, j int) bool {
			if positions[i].CompanyName == positions[j].CompanyName {
				return positions[i].CompanyID < positions[j].CompanyID
			}
			return positions[i].CompanyName < positions[j].CompanyName
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"progression": progression, "positions": positions})
}

func (h *Handler) stockOrder(w http.ResponseWriter, r *http.Request) {
	current, ok := h.authorizedPlayer(w, r)
	if !ok {
		return
	}
	var input stockOrderInput
	if !decodeJSON(w, r, &input) {
		return
	}
	input.ClientOrderID = strings.TrimSpace(input.ClientOrderID)
	input.CompanyID = strings.TrimSpace(input.CompanyID)
	input.Side = strings.ToLower(strings.TrimSpace(input.Side))
	city, exists := world.CityByID(h.worldFor(r.Context(), current.ID).CityID)
	if !exists {
		h.writeAPIError(w, r, http.StatusConflict, "PROXY_JURISDICTION_UNAVAILABLE", "Connect a Proxy in the company's country before placing an order.", nil)
		return
	}
	index, ok := h.worldReader.(companySiteIndexReader)
	if !ok {
		h.writeAPIError(w, r, http.StatusServiceUnavailable, "MARKET_LIST_UNAVAILABLE", "National company listings are temporarily unavailable.", nil)
		return
	}
	companies, err := index.CompanySitesInCountry(r.Context(), h.worldServerID, city.CountryID)
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "MARKET_LIST_FAILED", "National company listings could not be loaded.", err)
		return
	}
	var listing stockListing
	for _, record := range companies {
		if record.Company.ID == input.CompanyID {
			listing = stockListing{CompanyID: record.Company.ID, CompanyName: record.Company.Name, CountryID: city.CountryID, CityName: record.Address.City, PriceCents: stockQuoteCents(city.CountryID, record.Company.ID)}
			break
		}
	}
	if listing.CompanyID == "" {
		h.writeAPIError(w, r, http.StatusNotFound, "COMPANY_NOT_IN_ACTIVE_MARKET", "The company is not listed in the country of your active Proxy.", nil)
		return
	}
	writer, ok := h.store.(stockOrderWriter)
	if !ok {
		h.writeAPIError(w, r, http.StatusServiceUnavailable, "STOCK_TRADING_UNAVAILABLE", "Stock trading is temporarily unavailable.", nil)
		return
	}
	creditsCents, err := writer.executeStockOrder(r.Context(), current.ID, h.worldServerID, listing, input)
	switch {
	case errors.Is(err, ErrNoCharacter):
		h.writeAPIError(w, r, http.StatusConflict, "PLAYER_CHARACTER_REQUIRED", "Create a character before using the stock exchange.", nil)
		return
	case errors.Is(err, ErrInsufficientFunds):
		h.writeAPIError(w, r, http.StatusConflict, "INSUFFICIENT_CREDITS", "There are not enough Credits for this order.", nil)
		return
	case errors.Is(err, ErrInsufficientShares):
		h.writeAPIError(w, r, http.StatusConflict, "INSUFFICIENT_SHARES", "Your portfolio does not contain enough shares for this sale.", nil)
		return
	case errors.Is(err, ErrDuplicateStockOrder):
		h.writeAPIError(w, r, http.StatusConflict, "DUPLICATE_STOCK_ORDER", "This stock order has already been processed.", nil)
		return
	case errors.Is(err, ErrInvalidStockOrder):
		h.writeAPIError(w, r, http.StatusBadRequest, "INVALID_STOCK_ORDER", "Provide a valid side, quantity and request ID for the listed company.", nil)
		return
	case err != nil:
		h.writeAPIError(w, r, http.StatusInternalServerError, "STOCK_ORDER_FAILED", "The stock order could not be completed.", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"creditsCents": creditsCents, "order": map[string]any{"companyId": listing.CompanyID, "side": input.Side, "shares": input.Shares, "priceCents": listing.PriceCents, "totalCents": listing.PriceCents * input.Shares}})
}
