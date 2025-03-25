package api

import (
	"encoding/json"
	"net/http"

	"github.com/titi0001/Microservices-API-in-Go/api/dto"
	"github.com/titi0001/Microservices-API-in-Go/domain/ports"
	"github.com/titi0001/Microservices-API-in-Go/infrastructure/utils"
	"github.com/titi0001/Microservices-API-in-Go/logger"
)

type AuthHandler struct {
	service ports.AuthService
}

func NewAuthHandler(service ports.AuthService) *AuthHandler {
	if service == nil {
		logger.Fatal("AuthService cannot be nil", logger.String("component", "AuthHandler"))
		return nil
	}
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Warn("Invalid method for login", logger.String("method", r.Method))
		utils.WriteResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	var request dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Warn("Invalid login request payload", logger.Any("error", err))
		utils.WriteResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	response, appError := h.service.RemoteLogin(request)
	if appError != nil {
		logger.Warn("Login failed", logger.String("error", appError.Message))
		utils.WriteResponse(w, appError.Code, map[string]string{"error": appError.Message})
		return
	}
	utils.WriteResponse(w, http.StatusOK, response)
}

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	routeName := r.URL.Query().Get("routeName")
	vars := utils.ExtractQueryParams(r, []string{"token", "routeName"})

	isAuthorized, appError := h.service.RemoteIsAuthorized(token, routeName, vars)
	if appError != nil {
		logger.Warn("Authorization failed",
			logger.String("token", token),
			logger.String("routeName", routeName),
			logger.Any("error", appError))
		utils.WriteResponse(w, appError.Code, map[string]string{"error": appError.Message})
		return
	}

	claims, parseErr := utils.ExtractClaimsFromToken(token, h.service.GetSecretKey)
	if parseErr != nil {
		logger.Error("Failed to extract claims from token",
			logger.String("token", token),
			logger.Any("error", parseErr))
		utils.WriteResponse(w, http.StatusInternalServerError, map[string]string{"error": "Error processing token"})
		return
	}

	response := map[string]interface{}{
		"isAuthorized": isAuthorized,
		"role":         claims["role"],
	}
	statusCode := http.StatusOK
	if !isAuthorized {
		statusCode = http.StatusForbidden
	}
	utils.WriteResponse(w, statusCode, response)
}
