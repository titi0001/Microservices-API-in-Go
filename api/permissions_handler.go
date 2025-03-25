package api

import (
	"net/http"

	"github.com/titi0001/Microservices-API-in-Go/domain/ports"
	"github.com/titi0001/Microservices-API-in-Go/infrastructure/utils"
	"github.com/titi0001/Microservices-API-in-Go/logger"
)

type PermissionsHandler struct {
	service ports.AuthService
}

func NewPermissionsHandler(service ports.AuthService) *PermissionsHandler {
	return &PermissionsHandler{
		service: service,
	}
}

func (h *PermissionsHandler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		logger.Warn("Invalid method for GetRolePermissions", logger.String("method", r.Method))
		utils.WriteResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	rolePermissions := h.service.GetRolePermissions()
	response := map[string]interface{}{
		"permissions": rolePermissions.GetAllPermissions(),
	}

	utils.WriteResponse(w, http.StatusOK, response)
}
