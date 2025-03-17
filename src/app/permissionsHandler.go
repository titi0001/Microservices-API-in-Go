package app

import (
	"encoding/json"
	"net/http"

	"github.com/titi0001/Microservices-API-in-Go/src/domain"
	"github.com/titi0001/Microservices-API-in-Go/src/logger"
)

type PermissionsHandler struct {
	service domain.AuthService
}

func NewPermissionsHandler(service domain.AuthService) *PermissionsHandler {
	return &PermissionsHandler{
		service: service,
	}
}

func (h *PermissionsHandler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	rolePermissions := h.service.GetRolePermissions()

	response := map[string]interface{}{
		"permissions": rolePermissions.GetAllPermissions(),
	}

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Error encoding permissions", logger.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)

		errResponse := map[string]string{"error": "Internal error processing permissions"}
		if err := json.NewEncoder(w).Encode(errResponse); err != nil {
			logger.Error("Failed to encode error response", logger.Any("error", err))

			_, writeErr := w.Write([]byte(`{"error":"Internal error"}`))
			if writeErr != nil {
				logger.Error("Failed to write error response", logger.Any("error", writeErr))
			}
		}
	}
}
