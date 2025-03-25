package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/titi0001/Microservices-API-in-Go/domain"
	"github.com/titi0001/Microservices-API-in-Go/domain/ports"
	"github.com/titi0001/Microservices-API-in-Go/infrastructure/utils"
	"github.com/titi0001/Microservices-API-in-Go/logger"
)

const (
	TokenVerificationTimeout = 3 * time.Second
)

type AuthMiddleware struct {
	repo            ports.AuthRepository
	rolePermissions domain.RolePermissions
}

func NewAuthMiddleware(repo ports.AuthRepository) *AuthMiddleware {
	return &AuthMiddleware{
		repo:            repo,
		rolePermissions: *domain.GetRolePermissions(),
	}
}

func (a *AuthMiddleware) AuthorizationHandler() func(http.Handler) http.Handler {
	const (
		StatusUnauthorized        = http.StatusUnauthorized
		StatusForbidden           = http.StatusForbidden
		StatusInternalServerError = http.StatusInternalServerError
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			currentRoute := mux.CurrentRoute(r)
			if currentRoute == nil {
				logger.Error("No route matched for request")
				w.WriteHeader(http.StatusNotFound)
				return
			}

			currentRouteName := currentRoute.GetName()
			if currentRouteName == "AuthLogin" {
				logger.Info("Skipping auth for AuthLogin route")
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			logger.Debug("Authorization header", logger.String("header", authHeader))
			token := utils.GetTokenFromHeader(authHeader)
			if token == "" {
				logger.Warn("Missing or invalid auth token", logger.String("header", authHeader))
				utils.WriteResponse(w, StatusUnauthorized, map[string]string{"error": "Missing token"})
				return
			}

			currentRouteVars := mux.Vars(r)
			verifyURL := utils.BuildVerifyURL(token, currentRouteName, currentRouteVars)
			logger.Debug("Verification URL", logger.String("url", verifyURL))
			client := &http.Client{Timeout: TokenVerificationTimeout}

			resp, err := client.Get(verifyURL)
			if err != nil {
				logger.Error("Error verifying token", logger.Any("error", err))
				utils.WriteResponse(w, StatusInternalServerError, map[string]string{"error": "Error verifying token"})
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				logger.Warn("Token verification failed", logger.Int("status", resp.StatusCode))
				utils.WriteResponse(w, StatusUnauthorized, map[string]string{"error": "Unauthorized"})
				return
			}

			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.Error("Error reading response body", logger.Any("error", err))
				utils.WriteResponse(w, StatusInternalServerError, map[string]string{"error": "Error verifying token"})
				return
			}

			var verifyResponse map[string]interface{}
			if err := json.Unmarshal(responseBody, &verifyResponse); err != nil {
				logger.Error("Error parsing response", logger.Any("error", err))
				utils.WriteResponse(w, StatusInternalServerError, map[string]string{"error": "Error verifying token"})
				return
			}

			isAuthorized, ok := verifyResponse["isAuthorized"].(bool)
			if !ok || !isAuthorized {
				logger.Warn("Token not authorized", logger.Any("response", verifyResponse))
				utils.WriteResponse(w, StatusForbidden, map[string]string{"error": "Unauthorized"})
				return
			}

			userRole, ok := verifyResponse["role"].(string)
			if !ok {
				logger.Error("Role not found in response", logger.Any("response", verifyResponse))
				utils.WriteResponse(w, StatusInternalServerError, map[string]string{"error": "Invalid token claims"})
				return
			}

			if !a.rolePermissions.IsAuthorizedFor(userRole, currentRouteName) {
				logger.Warn("Insufficient permissions",
					logger.String("role", userRole),
					logger.String("routeName", currentRouteName))
				utils.WriteResponse(w, StatusForbidden, map[string]string{"error": "Insufficient permissions"})
				return
			}

			logger.Info("Authorization successful", logger.String("role", userRole), logger.String("route", currentRouteName))
			next.ServeHTTP(w, r)
		})
	}
}