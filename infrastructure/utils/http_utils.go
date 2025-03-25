package utils

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/titi0001/Microservices-API-in-Go/logger"
	"net/http"
	"net/url"
	"strings"
)

func WriteResponse(w http.ResponseWriter, code int, data any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode response", logger.Any("error", err))
	}
}

func BuildVerifyURL(token string, routeName string, vars map[string]string) string {
	u := url.URL{Host: "localhost:8181", Path: "/auth/verify", Scheme: "http"}
	q := u.Query()
	q.Add("token", token)
	q.Add("routeName", routeName)
	for k, v := range vars {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func ExtractQueryParams(r *http.Request, exclude []string) map[string]string {
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

func ExtractClaimsFromToken(tokenString string, secretKey func() []byte) (jwt.MapClaims, error) {
	logger.Info("Attempting to extract claims from token",
		logger.Int("token_length", len(tokenString)))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			method := "unknown"
			if alg, ok := token.Header["alg"].(string); ok {
				method = alg
			}
			logger.Error("Unexpected signing method", logger.String("alg", method))
			return nil, jwt.NewValidationError("unexpected signing method: "+method, jwt.ValidationErrorSignatureInvalid)
		}
		return secretKey(), nil
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
	return nil, jwt.NewValidationError("invalid token claims", jwt.ValidationErrorClaimsInvalid)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


func GetTokenFromHeader(header string) string {
	splitToken := strings.Split(header, "Bearer")
	if len(splitToken) == 2 {
		return strings.TrimSpace(splitToken[1])
	}
	return ""
}
