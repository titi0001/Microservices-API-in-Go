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
		"exp":         jwt.TimeFunc().Add(24 * time.Hour).Unix(), // Expira em 24 horas
	}


	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, signErr := token.SignedString(s.secretKey)
	if signErr != nil {
		logger.Error("Failed to generate JWT token", logger.Any("error", signErr))
		return nil, errs.NewUnexpectedError("Error generating token: " + signErr.Error())
	}

	return &dto.LoginResponse{Token: tokenString}, nil
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.LoginResponse, *errs.AppError) {
	user := domain.User{
		Username:   req.Username,
		Password:   req.Password,
		Role:       req.Role,
		CustomerID: req.CustomerID,
		CreatedOn:  time.Now().Format("2006-01-02 15:04:05"),
	}


	_, err := s.repo.SaveUser(user)
	if err != nil {
		logger.Error("Error saving new user", logger.Any("error", err))
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
		logger.Error("Failed to generate JWT token for new user", logger.Any("error", signErr))
		return nil, errs.NewUnexpectedError("Error generating token: " + signErr.Error())
	}

	return &dto.LoginResponse{Token: tokenString}, nil
}

func (s *AuthService) Refresh(token string) (*dto.LoginResponse, *errs.AppError) {

	claims, err := utils.ExtractClaimsFromToken(token, s.GetSecretKey)
	if err != nil {
		logger.Error("Failed to parse refresh token", logger.Any("error", err))
		return nil, errs.NewAuthenticationError("Invalid refresh token")
	}

	if exp, ok := claims["exp"].(float64); !ok || time.Unix(int64(exp), 0).Before(time.Now()) {
		logger.Warn("Refresh token expired")
		return nil, errs.NewAuthenticationError("Refresh token expired")
	}

	username, _ := claims["username"].(string)
	role, _ := claims["role"].(string)
	customerID, _ := claims["customer_id"].(string)

	newClaims := jwt.MapClaims{
		"username":    username,
		"role":        role,
		"customer_id": customerID,
		"exp":         jwt.TimeFunc().Add(24 * time.Hour).Unix(),
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenString, signErr := newToken.SignedString(s.secretKey)
	if signErr != nil {
		logger.Error("Failed to generate new JWT token for refresh", logger.Any("error", signErr))
		return nil, errs.NewUnexpectedError("Error generating new token: " + signErr.Error())
	}

	return &dto.LoginResponse{Token: newTokenString}, nil
}

func (s *AuthService) RemoteIsAuthorized(token, routeName string, vars map[string]string) (bool, *errs.AppError) {
	verifyURL := utils.BuildVerifyURL(token, routeName, vars)
	logger.Debug("Verification URL", logger.String("url", verifyURL))

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