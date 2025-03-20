package service

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/titi0001/Microservices-API-in-Go/src/domain"
	"github.com/titi0001/Microservices-API-in-Go/src/dto"
	"github.com/titi0001/Microservices-API-in-Go/src/errs"
	"github.com/titi0001/Microservices-API-in-Go/src/logger"
)

const (
	defaultJWTSecretKey = "secret"
	minPasswordLength   = 6
	tokenExpiryHours    = 24
)

type authService struct {
	authServerURL   string
	jwtSecretKey    []byte
	repository      domain.AuthRepository
	rolePermissions *domain.RolePermissions
}

func (s *authService) GetRolePermissions() *domain.RolePermissions {
	return s.rolePermissions
}

func (s *authService) GetSecretKey() []byte {
	return s.jwtSecretKey
}

func NewAuthService(authServerURL string, repository domain.AuthRepository) domain.AuthService {
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		jwtSecret = defaultJWTSecretKey
		logger.Warn("Using default JWT secret key. Consider setting JWT_SECRET_KEY environment variable")
	}

	return &authService{
		authServerURL:   authServerURL,
		jwtSecretKey:    []byte(jwtSecret),
		repository:      repository,
		rolePermissions: domain.GetRolePermissions(),
	}
}

func (s *authService) RemoteLogin(body io.Reader) ([]byte, *errs.AppError) {
	var req dto.LoginRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		logger.Error("Error decoding login request body", logger.Any("error", err))
		return nil, errs.NewBadRequestError("Invalid request body")
	}

	if err := s.validateLoginRequest(req.Username, req.Password); err != nil {
		return nil, err
	}

	user, err := s.repository.FindUser(req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	token, genErr := s.generateToken(req, user)
	if genErr != nil {
		logger.Error("Error generating token", logger.Any("error", genErr))
		return nil, errs.NewUnexpectedError("Failed to generate authentication token")
	}

	refreshToken, refreshErr := s.generateRefreshToken(req.Username)
	if refreshErr != nil {
		logger.Error("Error generating refresh", logger.Any("error", refreshErr))
		return nil, errs.NewUnexpectedError("Failed to generate refresh token")
	}

	if saveErr := s.repository.SaveRefreshToken(req.Username, refreshToken); saveErr != nil {
		logger.Error("Error saving refresh token to database", logger.Any("error", saveErr))
		return nil, errs.NewUnexpectedError("Failed to save refresh token")
	}

	respBody := dto.LoginResponse{
		Token: token,
		RefreshToken: refreshToken,
	}

	jsonResp, jsonErr := json.Marshal(respBody)
	if jsonErr != nil {
		logger.Error("Error encoding login response", logger.Any("error", jsonErr))
		return nil, errs.NewUnexpectedError("Error encoding response")
	}

	logger.Info("Successful login", logger.String("username", req.Username))
	return jsonResp, nil
}

func (s *authService) validateLoginRequest(username, password string) *errs.AppError {
	if username == "" || password == "" {
		logger.Warn("Missing required fields in login request",
			logger.String("username", username),
			logger.Bool("password_present", password != ""))
		return errs.NewBadRequestError("Username and password are required")
	}

	if len(password) < minPasswordLength {
		logger.Warn("Validation failed: insufficient password length",
			logger.String("username", username),
			logger.Int("password_length", len(password)))
		return errs.NewValidationError("Insufficient password length, minimum 6 characters required")
	}

	return nil
}

func (s *authService) generateToken(req dto.LoginRequest, user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"username":   req.Username,
		"role":       user.Role,
		"created_on": user.CreatedOn,
		"exp":        time.Now().Add(time.Hour * tokenExpiryHours).Unix(),
	}

	if user.CustomerId.Valid {
		claims["customer_id"] = user.CustomerId.Int64
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *authService) RemoteIsAuthorized(token string, routeName string, vars map[string]string) (bool, *errs.AppError) {
	if token == "" {
		logger.Warn("Empty token received")
		return false, errs.NewAuthenticationError("Token cannot be empty")
	}

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecretKey, nil
	})

	if err != nil {
		logger.Error("Error parsing JWT token", logger.String("token", token), logger.Any("error", err))
		return false, errs.NewAuthenticationError("Invalid token")
	}

	if !parsedToken.Valid {
		logger.Warn("Invalid or expired token", logger.String("token", token))
		return false, errs.NewAuthenticationError("Invalid or expired token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		logger.Error("Failed to extract claims from token")
		return false, errs.NewUnexpectedError("Error processing token claims")
	}

	role, ok := claims["role"].(string)
	if !ok {
		logger.Error("Role not found in token claims")
		return false, errs.NewUnexpectedError("Invalid token claims: role not found")
	}

	if !s.rolePermissions.IsAuthorizedFor(role, routeName) {
		logger.Warn("Unauthorized access attempt",
			logger.String("role", role),
			logger.String("routeName", routeName))
		return false, errs.NewForbiddenError("Insufficient permissions for route " + routeName)
	}

	customerId := claims["customer_id"]

	isAuthorized := s.repository.VerifyPermission(role, customerId, routeName, vars)
	if !isAuthorized {
		logger.Warn("Unauthorized access attempt",
			logger.String("role", role),
			logger.String("routeName", routeName))
		return false, errs.NewForbiddenError("Insufficient permissions for route " + routeName)
	}

	logger.Info("Access authorized",
		logger.String("routeName", routeName))
	return true, nil
}

func (s *authService) generateRefreshToken(username string) (string, error) {
	refreshClaims := jwt.MapClaims{
		"type":     "refresh_token",
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24 * 30).Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.jwtSecretKey)
	if err != nil {
		return "", err
	}
	return refreshTokenString, nil
}
