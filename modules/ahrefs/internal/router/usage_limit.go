package router

import (
	"math"
	"net/http"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/constant"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/repository"
)

type UsageLimitRouteHandler struct {
	Repository repository.Repository
}

func NewUsageLimitRouteHandler(repository repository.Repository) *UsageLimitRouteHandler {
	return &UsageLimitRouteHandler{
		Repository: repository,
	}
}

type GetUsageLimitResponse struct {
	UserID       int64 `json:"user_id"`
	ReportUsage  int   `json:"report_usage"`
	ExportUsage  int   `json:"export_usage"`
	LimitResetAt struct {
		Hour   int `json:"hour"`
		Minute int `json:"minute"`
	} `json:"limit_reset_at"`
}

func (h *UsageLimitRouteHandler) GetUsageLimit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// FIXME: enable this when cookie is available
	// sessionCookie, err := r.Cookie("PHPSESSID")
	// if err != nil {
	// 	log.Error().Err(err).Msg("failed to get session cookie")
	// }
	// if sessionCookie == nil {
	// 	w.WriteHeader(http.StatusForbidden)
	// 	w.Write([]byte("Access forbidden"))
	// 	return
	// }

	// userID, err := h.Repository.Session().FindUserIDBySession(ctx, sessionCookie.Value)
	// if err != nil {
	// 	log.Error().Err(err).Msg("failed to get user id from session cookie")

	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	w.Write([]byte("500 - Something bad happened!"))
	// 	return
	// }
	// if userID == 0 {
	// 	w.WriteHeader(http.StatusForbidden)
	// 	w.Write([]byte("Access forbidden"))
	// 	return
	// }

	var userID int64 = 5934

	usageLimit, err := h.Repository.UsageLimit().Retrieve(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get usage limit data")

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
		return
	}

	remainingSession := constant.UsageLimitDuration - usageLimit.LimitResetAt
	hour := math.Trunc(float64(remainingSession) / 60)
	minute := remainingSession - int(hour)*60

	response := &GetUsageLimitResponse{
		UserID:      userID,
		ReportUsage: usageLimit.ReportUsage,
		ExportUsage: usageLimit.ExportUsage,
		LimitResetAt: struct {
			Hour   int `json:"hour"`
			Minute int `json:"minute"`
		}{
			Hour:   int(hour),
			Minute: minute,
		},
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, response)
}
