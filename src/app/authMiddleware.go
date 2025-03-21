package app

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/titi0001/Microservices-API-in-Go/src/domain"
	"github.com/titi0001/Microservices-API-in-Go/src/errs"
	"github.com/titi0001/Microservices-API-in-Go/src/infrastructure/utils"
	"github.com/titi0001/Microservices-API-in-Go/src/logger"
)

const (
	TokenVerificationTimeout = 3 * time.Second
)

type AuthMiddleware struct {
	repo            domain.AuthRepository
	rolePermissions domain.RolePermissions
}

func NewAuthMiddleware(repo domain.AuthRepository) AuthMiddleware {
	return AuthMiddleware{
		repo:            repo,
		rolePermissions: *domain.GetRolePermissions(),
	}
}

func (a AuthMiddleware) authorizationHandler() func(http.Handler) http.Handler {
	const (
		StatusUnauthorized        = http.StatusUnauthorized
		StatusForbidden           = http.StatusForbidden
		StatusInternalServerError = http.StatusInternalServerError
		StatusServiceUnavailable  = http.StatusServiceUnavailable

		MsgUnauthorized             = "Unauthorized"
		MsgMissingToken             = "missing token"
		MsgTokenVerificationError   = "Error verifying token"
		MsgTokenVerificationTimeout = "Timeout verifying token"
		MsgAuthServerUnavailable    = "Authentication service unavailable"
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
				next.ServeHTTP(w, r)
				return
			}

			if !IsAuthServerRunning() {
				err := StartAuthServer()
				if err != nil {
					logger.Error("Failed to start auth server", logger.Any("error", err))
					appError := errs.AppError{Code: StatusServiceUnavailable, Message: MsgAuthServerUnavailable}
					utils.WriteResponse(w, appError.Code, appError.AsMessage())
					return
				}
				time.Sleep(200 * time.Millisecond)

				if !IsAuthServerRunning() {
					logger.Error("Auth server still not running after startup attempt")
					appError := errs.AppError{Code: StatusServiceUnavailable, Message: MsgAuthServerUnavailable}
					utils.WriteResponse(w, appError.Code, appError.AsMessage())
					return
				}
				logger.Info("Auth server started on demand")
			}

			currentRouteVars := mux.Vars(r)
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				logger.Error("Missing auth token")
				utils.WriteResponse(w, StatusUnauthorized, MsgMissingToken)
				return
			}

			token := getTokenFromHeader(authHeader)
			if token == "" {
				logger.Error("Invalid token format")
				utils.WriteResponse(w, StatusUnauthorized, MsgMissingToken)
				return
			}

			verifyURL := domain.BuildVerifyUrl(token, currentRouteName, currentRouteVars)
			client := &http.Client{
				Timeout: TokenVerificationTimeout,
			}

			resp, err := client.Get(verifyURL)
			if err != nil {
				logger.Error("Error verifying token", logger.Any("error", err))
				appError := errs.AppError{Code: StatusInternalServerError, Message: MsgTokenVerificationError}
				utils.WriteResponse(w, appError.Code, appError.AsMessage())
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				logger.Warn("Token verification failed", logger.Int("status", resp.StatusCode))
				appError := errs.AppError{Code: StatusUnauthorized, Message: MsgUnauthorized}
				utils.WriteResponse(w, appError.Code, appError.AsMessage())
				return
			}

			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.Error("Error reading response body", logger.Any("error", err))
				appError := errs.AppError{Code: StatusInternalServerError, Message: MsgTokenVerificationError}
				utils.WriteResponse(w, appError.Code, appError.AsMessage())
				return
			}

			var verifyResponse map[string]interface{}
			if err := json.Unmarshal(responseBody, &verifyResponse); err != nil {
				logger.Error("Error parsing response", logger.Any("error", err))
				appError := errs.AppError{Code: StatusInternalServerError, Message: MsgTokenVerificationError}
				utils.WriteResponse(w, appError.Code, appError.AsMessage())
				return
			}

			isAuthorized, ok := verifyResponse["isAuthorized"].(bool)
			if !ok || !isAuthorized {
				appError := errs.AppError{Code: StatusForbidden, Message: MsgUnauthorized}
				utils.WriteResponse(w, appError.Code, appError.AsMessage())
				return
			}

			var userRole string
			if role, ok := verifyResponse["role"].(string); ok {
				userRole = role
			} else if claims, ok := verifyResponse["claims"].(map[string]interface{}); ok {
				if role, ok := claims["role"].(string); ok {
					userRole = role
				} else {
					logger.Error("Role not found in token claims or response")
					appError := errs.AppError{Code: StatusInternalServerError, Message: "Invalid token claims"}
					utils.WriteResponse(w, appError.Code, appError.AsMessage())
					return
				}
			} else {
				logger.Error("No role or claims found in response")
				appError := errs.AppError{Code: StatusInternalServerError, Message: "Invalid response format"}
				utils.WriteResponse(w, appError.Code, appError.AsMessage())
				return
			}

			if !a.rolePermissions.IsAuthorizedFor(userRole, currentRouteName) {
				logger.Error("Unauthorized access",
					logger.String("role", userRole),
					logger.String("routeName", currentRouteName))
				appError := errs.AppError{Code: StatusForbidden, Message: "Insufficient permissions"}
				utils.WriteResponse(w, appError.Code, appError.AsMessage())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getTokenFromHeader(header string) string {
	splitToken := strings.Split(header, "Bearer")
	if len(splitToken) == 2 {
		return strings.TrimSpace(splitToken[1])
	}
	return ""
}
