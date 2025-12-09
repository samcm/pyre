package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/samcm/pyre/internal/backfill"
	"github.com/samcm/pyre/internal/polymarket"
	"github.com/samcm/pyre/internal/storage"
	"github.com/sirupsen/logrus"
)

// APIHandler implements the ServerInterface
type APIHandler struct {
	storage  storage.Storage
	sync     polymarket.Service
	backfill backfill.Service
	log      logrus.FieldLogger
}

var _ ServerInterface = (*APIHandler)(nil)

// NewHandler creates a new API handler
func NewHandler(
	storage storage.Storage,
	sync polymarket.Service,
	backfill backfill.Service,
	log logrus.FieldLogger,
) *APIHandler {
	return &APIHandler{
		storage:  storage,
		sync:     sync,
		backfill: backfill,
		log:      log.WithField("package", "api"),
	}
}

// GetLeaderboard returns the leaderboard of all users
func (h *APIHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request, params GetLeaderboardParams) {
	ctx := r.Context()

	sortBy := "totalPnl"
	if params.SortBy != nil {
		sortBy = string(*params.SortBy)
	}

	sortDirection := "desc"
	if params.SortDirection != nil {
		sortDirection = string(*params.SortDirection)
	}

	stats, err := h.storage.GetLeaderboard(ctx, sortBy, sortDirection)
	if err != nil {
		h.log.WithError(err).Error("failed to get leaderboard")
		respondError(w, http.StatusInternalServerError, "Failed to get leaderboard")
		return
	}

	// Sort leaderboard
	h.sortLeaderboard(stats, sortBy, sortDirection)

	// Convert to API response
	leaderboard := make([]LeaderboardEntry, len(stats))
	for i, stat := range stats {
		entry := LeaderboardEntry{
			Rank:          i + 1,
			Username:      stat.Username,
			TotalPnl:      stat.TotalPnl,
			RealizedPnl:   stat.RealizedPnl,
			UnrealizedPnl: stat.UnrealizedPnl,
		}
		if stat.OpenPositions > 0 {
			entry.OpenPositions = &stat.OpenPositions
		}
		if stat.WinRate > 0 {
			entry.WinRate = &stat.WinRate
		}
		if stat.ProfileImage != nil {
			entry.ProfileImage = stat.ProfileImage
		}

		// Get persona info for this user
		user, err := h.storage.GetUser(ctx, stat.Username)
		if err == nil {
			personaInfo, err := h.storage.GetUserPersonaInfo(ctx, user.ID)
			if err == nil && personaInfo != nil {
				entry.PersonaSlug = &personaInfo.Slug
				entry.PersonaDisplayName = &personaInfo.DisplayName
			}
		}

		leaderboard[i] = entry
	}

	respondJSON(w, http.StatusOK, leaderboard)
}

// TriggerSync triggers a manual sync
func (h *APIHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Trigger sync in background
	go func() {
		if err := h.sync.TriggerSync(ctx); err != nil {
			h.log.WithError(err).Error("sync failed")
		}
	}()

	w.WriteHeader(http.StatusAccepted)
}

// GetUsers returns all tracked users
func (h *APIHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dbUsers, err := h.storage.GetUsers(ctx)
	if err != nil {
		h.log.WithError(err).Error("failed to get users")
		respondError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	users := make([]User, 0, len(dbUsers))
	for _, dbUser := range dbUsers {
		addresses, err := h.storage.GetUserAddresses(ctx, dbUser.ID)
		if err != nil {
			h.log.WithError(err).WithField("user_id", dbUser.ID).Error("failed to get user addresses")
			continue
		}

		addressList := make([]string, len(addresses))
		for i, addr := range addresses {
			addressList[i] = addr.Address
		}

		user := User{
			Username:  dbUser.Username,
			Addresses: addressList,
		}
		if dbUser.LastSynced != nil {
			user.LastSynced = dbUser.LastSynced
		}
		if dbUser.ProfileImage != nil {
			user.ProfileImage = dbUser.ProfileImage
		}

		users = append(users, user)
	}

	respondJSON(w, http.StatusOK, users)
}

// GetUser returns details for a specific user
func (h *APIHandler) GetUser(w http.ResponseWriter, r *http.Request, username string) {
	ctx := r.Context()

	stats, err := h.storage.GetUserStats(ctx, username)
	if err != nil {
		h.log.WithError(err).WithField("username", username).Error("failed to get user stats")
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	detail := UserDetail{
		Username:      stats.Username,
		Addresses:     stats.Addresses,
		TotalPnl:      stats.TotalPnl,
		RealizedPnl:   stats.RealizedPnl,
		UnrealizedPnl: stats.UnrealizedPnl,
	}

	if stats.OpenPositions > 0 {
		detail.OpenPositions = &stats.OpenPositions
	}
	if stats.TotalTrades > 0 {
		detail.TotalTrades = &stats.TotalTrades
	}
	if stats.WinRate > 0 {
		detail.WinRate = &stats.WinRate
	}
	if stats.LastSynced != nil {
		detail.LastSynced = stats.LastSynced
	}
	if stats.ProfileImage != nil {
		detail.ProfileImage = stats.ProfileImage
	}

	respondJSON(w, http.StatusOK, detail)
}

// GetUserPnl returns PNL history for a user
func (h *APIHandler) GetUserPnl(w http.ResponseWriter, r *http.Request, username string, params GetUserPnlParams) {
	ctx := r.Context()

	user, err := h.storage.GetUser(ctx, username)
	if err != nil {
		h.log.WithError(err).WithField("username", username).Error("failed to get user")
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	var start, end *time.Time
	if params.Start != nil {
		start = params.Start
	}
	if params.End != nil {
		end = params.End
	}

	snapshots, err := h.storage.GetUserPnlHistory(ctx, user.ID, start, end)
	if err != nil {
		h.log.WithError(err).WithField("username", username).Error("failed to get pnl history")
		respondError(w, http.StatusInternalServerError, "Failed to get PNL history")
		return
	}

	dataPoints := make([]PnlDataPoint, len(snapshots))
	for i, snap := range snapshots {
		dataPoint := PnlDataPoint{
			Timestamp: snap.Timestamp,
		}
		if snap.TotalPnl != nil {
			dataPoint.TotalPnl = *snap.TotalPnl
		}
		if snap.RealizedPnl != nil {
			dataPoint.RealizedPnl = *snap.RealizedPnl
		}
		if snap.UnrealizedPnl != nil {
			dataPoint.UnrealizedPnl = *snap.UnrealizedPnl
		}
		dataPoints[i] = dataPoint
	}

	history := PnlHistory{
		Username:   username,
		DataPoints: dataPoints,
	}

	respondJSON(w, http.StatusOK, history)
}

// GetUserPositions returns current positions for a user
func (h *APIHandler) GetUserPositions(w http.ResponseWriter, r *http.Request, username string) {
	ctx := r.Context()

	user, err := h.storage.GetUser(ctx, username)
	if err != nil {
		h.log.WithError(err).WithField("username", username).Error("failed to get user")
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	dbPositions, err := h.storage.GetUserPositions(ctx, user.ID)
	if err != nil {
		h.log.WithError(err).WithField("username", username).Error("failed to get positions")
		respondError(w, http.StatusInternalServerError, "Failed to get positions")
		return
	}

	positions := make([]Position, 0, len(dbPositions))
	for _, pos := range dbPositions {
		position := Position{
			Id:            fmt.Sprintf("%d", pos.ID),
			MarketTitle:   "",
			Outcome:       "",
			Size:          0,
			AvgPrice:      0,
			CurrentPrice:  0,
			UnrealizedPnl: 0,
		}

		if pos.ConditionID != "" {
			position.ConditionId = &pos.ConditionID
		}
		if pos.MarketTitle != nil {
			position.MarketTitle = *pos.MarketTitle
		}
		if pos.MarketSlug != nil {
			position.MarketSlug = pos.MarketSlug
		}
		if pos.Outcome != nil {
			position.Outcome = *pos.Outcome
		}
		if pos.Size != nil {
			position.Size = *pos.Size
		}
		if pos.AvgPrice != nil {
			position.AvgPrice = *pos.AvgPrice
		}
		if pos.CurrentPrice != nil {
			position.CurrentPrice = *pos.CurrentPrice
		}
		if pos.InitialValue != nil {
			position.InitialValue = pos.InitialValue
		}
		if pos.CurrentValue != nil {
			position.CurrentValue = pos.CurrentValue
		}
		if pos.UnrealizedPnl != nil {
			position.UnrealizedPnl = *pos.UnrealizedPnl
		}
		if pos.UnrealizedPnlPercent != nil {
			position.UnrealizedPnlPercent = pos.UnrealizedPnlPercent
		}
		if pos.EndDate != nil {
			position.EndDate = pos.EndDate
		}

		positions = append(positions, position)
	}

	respondJSON(w, http.StatusOK, positions)
}

// GetUserTrades returns trade history for a user
func (h *APIHandler) GetUserTrades(w http.ResponseWriter, r *http.Request, username string, params GetUserTradesParams) {
	ctx := r.Context()

	user, err := h.storage.GetUser(ctx, username)
	if err != nil {
		h.log.WithError(err).WithField("username", username).Error("failed to get user")
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	limit := 100
	if params.Limit != nil {
		limit = *params.Limit
	}

	offset := 0
	if params.Offset != nil {
		offset = *params.Offset
	}

	dbTrades, total, err := h.storage.GetUserTrades(ctx, user.ID, limit, offset)
	if err != nil {
		h.log.WithError(err).WithField("username", username).Error("failed to get trades")
		respondError(w, http.StatusInternalServerError, "Failed to get trades")
		return
	}

	// Get persona info for this user (once, since all trades are from the same user)
	personaInfo, err := h.storage.GetUserPersonaInfo(ctx, user.ID)
	if err != nil {
		h.log.WithError(err).WithField("username", username).Error("failed to get persona info")
	}

	trades := make([]Trade, 0, len(dbTrades))
	for _, t := range dbTrades {
		trade := Trade{
			Id:          "",
			Timestamp:   time.Time{},
			MarketTitle: "",
			Outcome:     "",
			Side:        TradeSideBUY,
			Price:       0,
			Size:        0,
			Value:       0,
		}

		if t.TradeID != nil {
			trade.Id = *t.TradeID
		}
		if t.Timestamp != nil {
			trade.Timestamp = *t.Timestamp
		}
		if t.ConditionID != nil {
			trade.ConditionId = t.ConditionID
		}
		if t.MarketTitle != nil {
			trade.MarketTitle = *t.MarketTitle
		}
		if t.MarketSlug != nil {
			trade.MarketSlug = t.MarketSlug
		}
		if t.Outcome != nil {
			trade.Outcome = *t.Outcome
		}
		if t.Side != nil {
			if *t.Side == "BUY" {
				trade.Side = TradeSideBUY
			} else {
				trade.Side = TradeSideSELL
			}
		}
		if t.Price != nil {
			trade.Price = *t.Price
		}
		if t.Size != nil {
			trade.Size = *t.Size
		}
		if t.Value != nil {
			trade.Value = *t.Value
		}

		// Add persona info
		if personaInfo != nil {
			trade.PersonaSlug = &personaInfo.Slug
			trade.PersonaDisplayName = &personaInfo.DisplayName
		}

		// Add profile image
		if user.ProfileImage != nil {
			trade.ProfileImage = user.ProfileImage
		}

		trades = append(trades, trade)
	}

	response := TradesResponse{
		Trades: trades,
		Total:  total,
	}
	if limit > 0 {
		response.Limit = &limit
	}
	if offset > 0 {
		response.Offset = &offset
	}

	respondJSON(w, http.StatusOK, response)
}

// GetTrades returns all recent trades with filtering
func (h *APIHandler) GetTrades(w http.ResponseWriter, r *http.Request, params GetTradesParams) {
	ctx := r.Context()

	// Build filters from query parameters
	filters := storage.TradeFilters{
		Limit:         50,
		Offset:        0,
		SortBy:        "timestamp",
		SortDirection: "desc",
	}

	if params.Limit != nil {
		filters.Limit = *params.Limit
	}

	if params.Offset != nil {
		filters.Offset = *params.Offset
	}

	if params.Username != nil {
		filters.Username = params.Username
	}

	if params.Side != nil {
		side := string(*params.Side)
		filters.Side = &side
	}

	if params.MinValue != nil {
		filters.MinValue = params.MinValue
	}

	if params.SortBy != nil {
		filters.SortBy = string(*params.SortBy)
	}

	if params.SortDirection != nil {
		filters.SortDirection = string(*params.SortDirection)
	}

	dbTrades, total, err := h.storage.GetAllTrades(ctx, filters)
	if err != nil {
		h.log.WithError(err).Error("failed to get all trades")
		respondError(w, http.StatusInternalServerError, "Failed to get trades")
		return
	}

	// Cache for user lookups to avoid repeated queries
	userCache := make(map[int64]*storage.User, len(dbTrades))
	personaCache := make(map[int64]*storage.PersonaInfo, len(dbTrades))

	trades := make([]Trade, 0, len(dbTrades))
	for _, t := range dbTrades {
		trade := Trade{
			Id:          "",
			Timestamp:   time.Time{},
			MarketTitle: "",
			Outcome:     "",
			Side:        TradeSideBUY,
			Price:       0,
			Size:        0,
			Value:       0,
		}

		if t.TradeID != nil {
			trade.Id = *t.TradeID
		}
		if t.Username != "" {
			trade.Username = &t.Username
		}
		if t.Timestamp != nil {
			trade.Timestamp = *t.Timestamp
		}
		if t.ConditionID != nil {
			trade.ConditionId = t.ConditionID
		}
		if t.MarketTitle != nil {
			trade.MarketTitle = *t.MarketTitle
		}
		if t.MarketSlug != nil {
			trade.MarketSlug = t.MarketSlug
		}
		if t.Outcome != nil {
			trade.Outcome = *t.Outcome
		}
		if t.Side != nil {
			if *t.Side == "BUY" {
				trade.Side = TradeSideBUY
			} else {
				trade.Side = TradeSideSELL
			}
		}
		if t.Price != nil {
			trade.Price = *t.Price
		}
		if t.Size != nil {
			trade.Size = *t.Size
		}
		if t.Value != nil {
			trade.Value = *t.Value
		}

		// Get user info (with caching)
		user, ok := userCache[t.UserID]
		if !ok {
			user, err = h.storage.GetUser(ctx, t.Username)
			if err == nil {
				userCache[t.UserID] = user
			}
		}

		// Add profile image
		if user != nil && user.ProfileImage != nil {
			trade.ProfileImage = user.ProfileImage
		}

		// Get persona info (with caching)
		personaInfo, ok := personaCache[t.UserID]
		if !ok {
			personaInfo, err = h.storage.GetUserPersonaInfo(ctx, t.UserID)
			if err == nil {
				personaCache[t.UserID] = personaInfo
			}
		}

		// Add persona info
		if personaInfo != nil {
			trade.PersonaSlug = &personaInfo.Slug
			trade.PersonaDisplayName = &personaInfo.DisplayName
		}

		trades = append(trades, trade)
	}

	response := TradesResponse{
		Trades: trades,
		Total:  total,
	}
	if filters.Limit > 0 {
		response.Limit = &filters.Limit
	}
	if filters.Offset > 0 {
		response.Offset = &filters.Offset
	}

	respondJSON(w, http.StatusOK, response)
}

// sortLeaderboard sorts the leaderboard by the specified field and direction
func (h *APIHandler) sortLeaderboard(stats []*storage.UserStats, sortBy, sortDirection string) {
	sort.Slice(stats, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "totalPnl":
			less = stats[i].TotalPnl < stats[j].TotalPnl
		case "realizedPnl":
			less = stats[i].RealizedPnl < stats[j].RealizedPnl
		case "unrealizedPnl":
			less = stats[i].UnrealizedPnl < stats[j].UnrealizedPnl
		case "winRate":
			less = stats[i].WinRate < stats[j].WinRate
		default:
			less = stats[i].TotalPnl < stats[j].TotalPnl
		}

		if sortDirection == "asc" {
			return less
		}
		return !less
	})
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// parseIntParam parses an integer query parameter
func parseIntParam(r *http.Request, param string, defaultValue int) int {
	value := r.URL.Query().Get(param)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// BackfillUserPnl backfills PnL history from trade data for a user
func (h *APIHandler) BackfillUserPnl(w http.ResponseWriter, r *http.Request, username string) {
	ctx := r.Context()

	h.log.WithField("username", username).Info("starting PnL backfill")

	result, err := h.backfill.BackfillUser(ctx, username)
	if err != nil {
		h.log.WithError(err).WithField("username", username).Error("failed to backfill PnL")

		// Check if it's a user not found error
		if err.Error() == fmt.Sprintf("failed to get user: user not found: %s", username) {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}

		respondError(w, http.StatusInternalServerError, "Failed to backfill PnL")
		return
	}

	response := BackfillResult{
		Username:         result.Username,
		TradesProcessed:  result.TradesProcessed,
		SnapshotsCreated: result.SnapshotsCreated,
		TotalRealizedPnl: result.TotalRealizedPnl,
	}

	if result.OldestTradeDate != nil {
		response.OldestTradeDate = result.OldestTradeDate
	}
	if result.NewestTradeDate != nil {
		response.NewestTradeDate = result.NewestTradeDate
	}

	respondJSON(w, http.StatusOK, response)
}

// GetPersonas returns all personas
func (h *APIHandler) GetPersonas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dbPersonas, err := h.storage.GetPersonas(ctx)
	if err != nil {
		h.log.WithError(err).Error("failed to get personas")
		respondError(w, http.StatusInternalServerError, "Failed to get personas")
		return
	}

	personas := make([]PersonaSummary, 0, len(dbPersonas))
	for _, p := range dbPersonas {
		// Get users for this persona
		users, err := h.storage.GetPersonaUsers(ctx, p.ID)
		if err != nil {
			h.log.WithError(err).WithField("persona", p.Slug).Error("failed to get persona users")
			continue
		}

		usernames := make([]string, len(users))
		for i, u := range users {
			usernames[i] = u.Username
		}

		summary := PersonaSummary{
			Slug:        p.Slug,
			DisplayName: p.DisplayName,
			Usernames:   usernames,
		}
		if p.Image != nil {
			summary.Image = p.Image
		}
		personas = append(personas, summary)
	}

	respondJSON(w, http.StatusOK, personas)
}

// GetPersona returns details for a specific persona
func (h *APIHandler) GetPersona(w http.ResponseWriter, r *http.Request, slug string) {
	ctx := r.Context()

	stats, err := h.storage.GetPersonaStats(ctx, slug)
	if err != nil {
		h.log.WithError(err).WithField("slug", slug).Error("failed to get persona stats")
		respondError(w, http.StatusNotFound, "Persona not found")
		return
	}

	detail := PersonaDetail{
		Slug:          stats.Slug,
		DisplayName:   stats.DisplayName,
		Usernames:     stats.Usernames,
		TotalPnl:      stats.TotalPnl,
		RealizedPnl:   stats.RealizedPnl,
		UnrealizedPnl: stats.UnrealizedPnl,
	}

	if stats.OpenPositions > 0 {
		detail.OpenPositions = &stats.OpenPositions
	}
	if stats.TotalTrades > 0 {
		detail.TotalTrades = &stats.TotalTrades
	}
	if stats.WinRate > 0 {
		detail.WinRate = &stats.WinRate
	}
	if stats.Image != nil {
		detail.Image = stats.Image
	}

	respondJSON(w, http.StatusOK, detail)
}

// GetPersonaAccounts returns all accounts for a persona with individual stats
func (h *APIHandler) GetPersonaAccounts(w http.ResponseWriter, r *http.Request, slug string) {
	ctx := r.Context()

	persona, err := h.storage.GetPersona(ctx, slug)
	if err != nil {
		h.log.WithError(err).WithField("slug", slug).Error("failed to get persona")
		respondError(w, http.StatusNotFound, "Persona not found")
		return
	}

	users, err := h.storage.GetPersonaUsers(ctx, persona.ID)
	if err != nil {
		h.log.WithError(err).WithField("slug", slug).Error("failed to get persona users")
		respondError(w, http.StatusInternalServerError, "Failed to get persona accounts")
		return
	}

	accounts := make([]PersonaAccount, 0, len(users))
	for _, user := range users {
		stats, err := h.storage.GetUserStats(ctx, user.Username)
		if err != nil {
			h.log.WithError(err).WithField("username", user.Username).Error("failed to get user stats")
			continue
		}

		account := PersonaAccount{
			Username:      stats.Username,
			Addresses:     stats.Addresses,
			TotalPnl:      stats.TotalPnl,
			RealizedPnl:   stats.RealizedPnl,
			UnrealizedPnl: stats.UnrealizedPnl,
		}

		if stats.OpenPositions > 0 {
			account.OpenPositions = &stats.OpenPositions
		}
		if stats.TotalTrades > 0 {
			account.TotalTrades = &stats.TotalTrades
		}
		if stats.WinRate > 0 {
			account.WinRate = &stats.WinRate
		}
		if stats.ProfileImage != nil {
			account.ProfileImage = stats.ProfileImage
		}

		accounts = append(accounts, account)
	}

	respondJSON(w, http.StatusOK, accounts)
}

// GetPersonaLeaderboard returns the leaderboard of all personas
func (h *APIHandler) GetPersonaLeaderboard(w http.ResponseWriter, r *http.Request, params GetPersonaLeaderboardParams) {
	ctx := r.Context()

	sortBy := "totalPnl"
	if params.SortBy != nil {
		sortBy = string(*params.SortBy)
	}

	sortDirection := "desc"
	if params.SortDirection != nil {
		sortDirection = string(*params.SortDirection)
	}

	stats, err := h.storage.GetPersonaLeaderboard(ctx, sortBy, sortDirection)
	if err != nil {
		h.log.WithError(err).Error("failed to get persona leaderboard")
		respondError(w, http.StatusInternalServerError, "Failed to get persona leaderboard")
		return
	}

	// Sort leaderboard
	h.sortPersonaLeaderboard(stats, sortBy, sortDirection)

	// Convert to API response
	leaderboard := make([]PersonaLeaderboardEntry, len(stats))
	for i, stat := range stats {
		entry := PersonaLeaderboardEntry{
			Rank:          i + 1,
			Slug:          stat.Slug,
			DisplayName:   stat.DisplayName,
			Usernames:     &stat.Usernames,
			TotalPnl:      stat.TotalPnl,
			RealizedPnl:   stat.RealizedPnl,
			UnrealizedPnl: stat.UnrealizedPnl,
		}
		if stat.OpenPositions > 0 {
			entry.OpenPositions = &stat.OpenPositions
		}
		if stat.WinRate > 0 {
			entry.WinRate = &stat.WinRate
		}
		if stat.Image != nil {
			entry.Image = stat.Image
		}
		leaderboard[i] = entry
	}

	respondJSON(w, http.StatusOK, leaderboard)
}

// sortPersonaLeaderboard sorts the persona leaderboard by the specified field and direction
func (h *APIHandler) sortPersonaLeaderboard(stats []*storage.PersonaStats, sortBy, sortDirection string) {
	sort.Slice(stats, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "totalPnl":
			less = stats[i].TotalPnl < stats[j].TotalPnl
		case "realizedPnl":
			less = stats[i].RealizedPnl < stats[j].RealizedPnl
		case "unrealizedPnl":
			less = stats[i].UnrealizedPnl < stats[j].UnrealizedPnl
		case "winRate":
			less = stats[i].WinRate < stats[j].WinRate
		default:
			less = stats[i].TotalPnl < stats[j].TotalPnl
		}

		if sortDirection == "asc" {
			return less
		}
		return !less
	})
}

// GetPersonaPositions returns combined positions across all accounts for a persona
func (h *APIHandler) GetPersonaPositions(w http.ResponseWriter, r *http.Request, slug string) {
	ctx := r.Context()

	dbPositions, err := h.storage.GetPersonaPositions(ctx, slug)
	if err != nil {
		h.log.WithError(err).WithField("slug", slug).Error("failed to get persona positions")
		respondError(w, http.StatusNotFound, "Persona not found")
		return
	}

	positions := make([]PersonaPosition, 0, len(dbPositions))
	for _, pos := range dbPositions {
		position := PersonaPosition{
			Id:            fmt.Sprintf("%d", pos.ID),
			Username:      pos.Username,
			MarketTitle:   "",
			Outcome:       "",
			Size:          0,
			AvgPrice:      0,
			CurrentPrice:  0,
			UnrealizedPnl: 0,
		}

		if pos.ConditionID != "" {
			position.ConditionId = &pos.ConditionID
		}
		if pos.MarketTitle != nil {
			position.MarketTitle = *pos.MarketTitle
		}
		if pos.MarketSlug != nil {
			position.MarketSlug = pos.MarketSlug
		}
		if pos.Outcome != nil {
			position.Outcome = *pos.Outcome
		}
		if pos.Size != nil {
			position.Size = *pos.Size
		}
		if pos.AvgPrice != nil {
			position.AvgPrice = *pos.AvgPrice
		}
		if pos.CurrentPrice != nil {
			position.CurrentPrice = *pos.CurrentPrice
		}
		if pos.InitialValue != nil {
			position.InitialValue = pos.InitialValue
		}
		if pos.CurrentValue != nil {
			position.CurrentValue = pos.CurrentValue
		}
		if pos.UnrealizedPnl != nil {
			position.UnrealizedPnl = *pos.UnrealizedPnl
		}
		if pos.UnrealizedPnlPercent != nil {
			position.UnrealizedPnlPercent = pos.UnrealizedPnlPercent
		}
		if pos.EndDate != nil {
			position.EndDate = pos.EndDate
		}

		positions = append(positions, position)
	}

	respondJSON(w, http.StatusOK, positions)
}

// GetPersonaTrades returns combined trades across all accounts for a persona
func (h *APIHandler) GetPersonaTrades(w http.ResponseWriter, r *http.Request, slug string, params GetPersonaTradesParams) {
	ctx := r.Context()

	limit := 100
	if params.Limit != nil {
		limit = *params.Limit
	}

	offset := 0
	if params.Offset != nil {
		offset = *params.Offset
	}

	// Get persona info upfront (all trades will share this)
	persona, err := h.storage.GetPersona(ctx, slug)
	if err != nil {
		h.log.WithError(err).WithField("slug", slug).Error("failed to get persona")
		respondError(w, http.StatusNotFound, "Persona not found")
		return
	}

	dbTrades, total, err := h.storage.GetPersonaTrades(ctx, slug, limit, offset)
	if err != nil {
		h.log.WithError(err).WithField("slug", slug).Error("failed to get persona trades")
		respondError(w, http.StatusNotFound, "Persona not found")
		return
	}

	// Cache for user lookups to avoid repeated queries
	userCache := make(map[int64]*storage.User, len(dbTrades))

	trades := make([]Trade, 0, len(dbTrades))
	for _, t := range dbTrades {
		trade := Trade{
			Id:          "",
			Timestamp:   time.Time{},
			MarketTitle: "",
			Outcome:     "",
			Side:        TradeSideBUY,
			Price:       0,
			Size:        0,
			Value:       0,
		}

		if t.TradeID != nil {
			trade.Id = *t.TradeID
		}
		trade.Username = &t.Username
		if t.Timestamp != nil {
			trade.Timestamp = *t.Timestamp
		}
		if t.ConditionID != nil {
			trade.ConditionId = t.ConditionID
		}
		if t.MarketTitle != nil {
			trade.MarketTitle = *t.MarketTitle
		}
		if t.MarketSlug != nil {
			trade.MarketSlug = t.MarketSlug
		}
		if t.Outcome != nil {
			trade.Outcome = *t.Outcome
		}
		if t.Side != nil {
			if *t.Side == "BUY" {
				trade.Side = TradeSideBUY
			} else {
				trade.Side = TradeSideSELL
			}
		}
		if t.Price != nil {
			trade.Price = *t.Price
		}
		if t.Size != nil {
			trade.Size = *t.Size
		}
		if t.Value != nil {
			trade.Value = *t.Value
		}

		// Get user info (with caching)
		user, ok := userCache[t.UserID]
		if !ok {
			user, err = h.storage.GetUser(ctx, t.Username)
			if err == nil {
				userCache[t.UserID] = user
			}
		}

		// Add profile image
		if user != nil && user.ProfileImage != nil {
			trade.ProfileImage = user.ProfileImage
		}

		// Add persona info (same for all trades in this response)
		trade.PersonaSlug = &persona.Slug
		trade.PersonaDisplayName = &persona.DisplayName

		trades = append(trades, trade)
	}

	response := TradesResponse{
		Trades: trades,
		Total:  total,
	}
	if limit > 0 {
		response.Limit = &limit
	}
	if offset > 0 {
		response.Offset = &offset
	}

	respondJSON(w, http.StatusOK, response)
}

// GetUserResults returns resolved positions (results) for a user
func (h *APIHandler) GetUserResults(w http.ResponseWriter, r *http.Request, username string, params GetUserResultsParams) {
	ctx := r.Context()

	user, err := h.storage.GetUser(ctx, username)
	if err != nil {
		h.log.WithError(err).WithField("username", username).Error("failed to get user")
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	limit := 50
	if params.Limit != nil {
		limit = *params.Limit
	}

	offset := 0
	if params.Offset != nil {
		offset = *params.Offset
	}

	dbResults, total, err := h.storage.GetUserResults(ctx, user.ID, limit, offset)
	if err != nil {
		h.log.WithError(err).WithField("username", username).Error("failed to get results")
		respondError(w, http.StatusInternalServerError, "Failed to get results")
		return
	}

	results := make([]Result, 0, len(dbResults))
	for _, r := range dbResults {
		result := Result{
			Id:          fmt.Sprintf("%d", r.ID),
			ConditionId: r.ConditionID,
			MarketTitle: "",
			Outcome:     "",
			RealizedPnl: r.RealizedPnl,
		}

		if r.MarketTitle != nil {
			result.MarketTitle = *r.MarketTitle
		}
		if r.MarketSlug != nil {
			result.MarketSlug = r.MarketSlug
		}
		if r.Outcome != nil {
			result.Outcome = *r.Outcome
		}
		if r.InitialValue != nil {
			result.InitialValue = r.InitialValue
		}
		if r.EndDate != nil {
			result.EndDate = r.EndDate
		}
		if r.ResolutionDate != nil {
			result.ResolutionDate = r.ResolutionDate
		}

		results = append(results, result)
	}

	response := ResultsResponse{
		Results: results,
		Total:   total,
	}
	if limit > 0 {
		response.Limit = &limit
	}
	if offset > 0 {
		response.Offset = &offset
	}

	respondJSON(w, http.StatusOK, response)
}

// GetPersonaResults returns resolved positions (results) across all accounts for a persona
func (h *APIHandler) GetPersonaResults(w http.ResponseWriter, r *http.Request, slug string, params GetPersonaResultsParams) {
	ctx := r.Context()

	limit := 50
	if params.Limit != nil {
		limit = *params.Limit
	}

	offset := 0
	if params.Offset != nil {
		offset = *params.Offset
	}

	dbResults, total, err := h.storage.GetPersonaResults(ctx, slug, limit, offset)
	if err != nil {
		h.log.WithError(err).WithField("slug", slug).Error("failed to get persona results")
		respondError(w, http.StatusNotFound, "Persona not found")
		return
	}

	results := make([]PersonaResult, 0, len(dbResults))
	for _, r := range dbResults {
		result := PersonaResult{
			Id:          fmt.Sprintf("%d", r.ID),
			Username:    r.Username,
			ConditionId: r.ConditionID,
			MarketTitle: "",
			Outcome:     "",
			RealizedPnl: r.RealizedPnl,
		}

		if r.MarketTitle != nil {
			result.MarketTitle = *r.MarketTitle
		}
		if r.MarketSlug != nil {
			result.MarketSlug = r.MarketSlug
		}
		if r.Outcome != nil {
			result.Outcome = *r.Outcome
		}
		if r.InitialValue != nil {
			result.InitialValue = r.InitialValue
		}
		if r.EndDate != nil {
			result.EndDate = r.EndDate
		}
		if r.ResolutionDate != nil {
			result.ResolutionDate = r.ResolutionDate
		}

		results = append(results, result)
	}

	response := PersonaResultsResponse{
		Results: results,
		Total:   total,
	}
	if limit > 0 {
		response.Limit = &limit
	}
	if offset > 0 {
		response.Offset = &offset
	}

	respondJSON(w, http.StatusOK, response)
}
