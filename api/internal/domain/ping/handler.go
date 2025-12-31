package ping

import (
	"context"
	"net/http"
	"strings"
	"time"

	"base/api/pkg/response"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get client IP
	ip := getClientIP(r)

	// Save to both databases
	if err := h.repo.UpsertIPPostgres(ctx, ip); err != nil {
		response.InternalError(w, "failed to save to postgres: "+err.Error())
		return
	}

	if err := h.repo.InsertPingDynamo(ctx, ip); err != nil {
		response.InternalError(w, "failed to save to dynamo: "+err.Error())
		return
	}

	// Get last 5 from both databases
	postgresIPs, err := h.repo.GetLastIPsPostgres(ctx, 5)
	if err != nil {
		response.InternalError(w, "failed to get postgres ips: "+err.Error())
		return
	}

	dynamoPings, err := h.repo.GetLastPingsDynamo(ctx, 5)
	if err != nil {
		response.InternalError(w, "failed to get dynamo pings: "+err.Error())
		return
	}

	response.OK(w, PingResponse{
		YourIP:      ip,
		PostgresIPs: postgresIPs,
		DynamoPings: dynamoPings,
	})
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}
