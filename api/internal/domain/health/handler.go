package health

import (
	"context"
	"net/http"
	"time"

	"base/api/internal/database"
	"base/api/pkg/response"
)

type Handler struct {
	postgres *database.PostgresDB
	dynamo   *database.DynamoDB
	redis    *database.RedisDB
}

func NewHandler(postgres *database.PostgresDB, dynamo *database.DynamoDB, redis *database.RedisDB) *Handler {
	return &Handler{
		postgres: postgres,
		dynamo:   dynamo,
		redis:    redis,
	}
}

type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	response.OK(w, map[string]string{"status": "ok"})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)
	allHealthy := true

	// Check PostgreSQL
	if h.postgres != nil {
		if err := h.postgres.Health(ctx); err != nil {
			services["postgres"] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			services["postgres"] = "healthy"
		}
	}

	// Check DynamoDB
	if h.dynamo != nil {
		if err := h.dynamo.Health(ctx); err != nil {
			services["dynamodb"] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			services["dynamodb"] = "healthy"
		}
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Health(ctx); err != nil {
			services["redis"] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			services["redis"] = "healthy"
		}
	}

	status := "ok"
	httpStatus := http.StatusOK
	if !allHealthy {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	response.JSON(w, httpStatus, HealthResponse{
		Status:   status,
		Services: services,
	})
}
