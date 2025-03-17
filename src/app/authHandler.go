package app

import (
	"errors"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/titi0001/Microservices-API-in-Go/src/domain"
	"github.com/titi0001/Microservices-API-in-Go/src/infrastructure/utils"
	"github.com/titi0001/Microservices-API-in-Go/src/logger"
)

type AuthHandler struct {
	service domain.AuthService
}

func NewAuthHandler(service domain.AuthService) *AuthHandler {
	if service == nil {
		logger.Fatal("AuthService cannot be nil", logger.String("component", "AuthHandler"))
		return nil
	}
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.methodNotAllowed(w, r, "login")
		return
	}

	jsonResponse, err := h.service.RemoteLogin(r.Body)
	if err != nil {
		logger.Warn("Login failed", logger.String("error", err.Message))
		h.respondWithError(w, err.Message, err.Code)
		return
	}
	h.respondWithJSON(w, http.StatusOK, jsonResponse)
}

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	routeName := r.URL.Query().Get("routeName")

	vars := h.extractQueryParams(r, []string{"token", "routeName"})

	isAuthorized, err := h.service.RemoteIsAuthorized(token, routeName, vars)
	if err != nil {
		logger.Warn("Authorization failed",
			logger.String("token", token),
			logger.String("routeName", routeName),
			logger.Any("error", err))

		h.respondWithError(w, err.Message, err.Code)
		return
	}

	claims, parseErr := h.extractClaimsFromToken(token)
	if parseErr != nil {
		logger.Error("Failed to extract claims from token", 
			logger.String("token", token),
			logger.Any("error", parseErr))
		h.respondWithError(w, "Error processing token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"isAuthorized": isAuthorized,
		"role": claims["role"], 
	}
	
	statusCode := http.StatusOK
	if !isAuthorized {
		statusCode = http.StatusForbidden
	}

	utils.WriteResponse(w, statusCode, response)
}

func (h *AuthHandler) extractQueryParams(r *http.Request, exclude []string) map[string]string {
	vars := make(map[string]string)
	excludeMap := make(map[string]bool)

	for _, key := range exclude {
		excludeMap[key] = true
	}

	for key, values := range r.URL.Query() {
		if !excludeMap[key] && len(values) > 0 {
			vars[key] = values[0]
		}
	}

	return vars
}

func (h *AuthHandler) methodNotAllowed(w http.ResponseWriter, r *http.Request, endpoint string) {
	logger.Warn("Invalid method for endpoint",
		logger.String("endpoint", endpoint),
		logger.String("method", r.Method))
	h.respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *AuthHandler) respondWithError(w http.ResponseWriter, message string, statusCode int) {
	utils.WriteResponse(w, statusCode, map[string]string{"error": message})
}

func (h *AuthHandler) respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	if jsonBytes, ok := data.([]byte); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, err := w.Write(jsonBytes)
		if err != nil {
			logger.Error("Failed to write response", logger.Any("error", err))
		}
		return
	}

	utils.WriteResponse(w, statusCode, data)
}

func (h *AuthHandler) extractClaimsFromToken(tokenString string) (jwt.MapClaims, error) {
    logger.Info("Attempting to extract claims from token",
        logger.Int("token_length", len(tokenString)))
    
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            method := "unknown"
            if alg, ok := token.Header["alg"].(string); ok {
                method = alg
            }
            logger.Error("Unexpected signing method",
                logger.String("alg", method))
            return nil, errors.New("unexpected signing method: " + method)
        }
        return h.service.GetSecretKey(), nil
    })

    if err != nil {
        prefix := ""
        if len(tokenString) > 0 {
            endIndex := min(10, len(tokenString))
            prefix = tokenString[:endIndex]
        }
        logger.Error("Failed to parse token",
            logger.String("token_prefix", prefix),
            logger.Any("error", err))
        return nil, err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok {
        claimKeys := make([]string, 0, len(claims))
        for k := range claims {
            claimKeys = append(claimKeys, k)
        }
        logger.Info("Successfully extracted claims from token",
            logger.Any("available_claims", claimKeys))
        return claims, nil
    }

    logger.Error("Invalid token claims format")
    return nil, errors.New("invalid token claims")
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}