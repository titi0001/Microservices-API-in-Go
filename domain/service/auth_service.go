package service

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/titi0001/Microservices-API-in-Go/api/dto"
	"github.com/titi0001/Microservices-API-in-Go/domain"
	"github.com/titi0001/Microservices-API-in-Go/domain/ports"
	"github.com/titi0001/Microservices-API-in-Go/errs"
	"github.com/titi0001/Microservices-API-in-Go/infrastructure/utils"
	"github.com/titi0001/Microservices-API-in-Go/logger"
)

type AuthService struct {
	serviceURL string
	repo       ports.AuthRepository
	secretKey  []byte
}


func NewAuthService(serviceURL string, repo ports.AuthRepository) *AuthService {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		logger.Fatal("JWT_SECRET_KEY environment variable not set")
	}

	return &AuthService{
		serviceURL: serviceURL,
		repo:       repo,
		secretKey:  []byte(secretKey),
	}
}

func (s *AuthService) RemoteLogin(req dto.LoginRequest) (*dto.LoginResponse, *errs.AppError) {
	user, err := s.repo.FindUser(req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	claims := jwt.MapClaims{
		"username":    user.Username,
		"role":        user.Role,
		"customer_id": user.CustomerID,
		"exp":         jwt.TimeFunc().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, signErr := token.SignedString(s.secretKey)
	if signErr != nil {
		logger.Error("Failed to generate JWT token", logger.Any("error", signErr))
		return nil, errs.NewUnexpectedError("Error generating token: " + signErr.Error())
	}

	return &dto.LoginResponse{Token: tokenString}, nil
}

func (s *AuthService) RemoteIsAuthorized(token, routeName string, vars map[string]string) (bool, *errs.AppError) {

	verifyURL := utils.BuildVerifyURL(token, routeName, vars)

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(verifyURL)
	if err != nil {
		logger.Error("Error verifying token remotely", logger.Any("error", err))
		return false, errs.NewUnexpectedError("Error verifying token")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warn("Token verification failed", logger.Int("status", resp.StatusCode))
		return false, errs.NewAuthenticationError("Unauthorized")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading verification response", logger.Any("error", err))
		return false, errs.NewUnexpectedError("Error reading response")
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		logger.Error("Error parsing verification response", logger.Any("error", err))
		return false, errs.NewUnexpectedError("Error parsing response")
	}

	isAuthorized, ok := response["isAuthorized"].(bool)
	if !ok {
		logger.Error("Invalid verification response format", logger.Any("response", response))
		return false, errs.NewUnexpectedError("Invalid response format")
	}

	return isAuthorized, nil
}

func (s *AuthService) GetSecretKey() []byte {
	return s.secretKey
}

func (s *AuthService) GetRolePermissions() domain.RolePermissions {
	return *domain.GetRolePermissions()
}