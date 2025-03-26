package service

import (
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

	customerIDClaim := ""
	if user.CustomerID != nil {
		customerIDClaim = *user.CustomerID
	}

	claims := jwt.MapClaims{
		"username":    user.Username,
		"role":        user.Role,
		"customer_id": customerIDClaim,
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

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.LoginResponse, *errs.AppError) {
	var customerID *string
	if req.CustomerID != "" {
		customerID = &req.CustomerID
	}

	user := domain.User{
		Username:   req.Username,
		Password:   req.Password,
		Role:       req.Role,
		CustomerID: customerID,
		CreatedOn:  time.Now(),
	}

	_, err := s.repo.SaveUser(user)
	if err != nil {
		logger.Error("Error saving new user", logger.Any("error", err))
		return nil, err
	}

	customerIDClaim := ""
	if user.CustomerID != nil {
		customerIDClaim = *user.CustomerID
	}

	claims := jwt.MapClaims{
		"username":    user.Username,
		"role":        user.Role,
		"customer_id": customerIDClaim,
		"exp":         jwt.TimeFunc().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, signErr := token.SignedString(s.secretKey)
	if signErr != nil {
		logger.Error("Failed to generate JWT token for new user", logger.Any("error", signErr))
		return nil, errs.NewUnexpectedError("Error generating token: " + signErr.Error())
	}

	refreshClaims := jwt.MapClaims{
		"username": user.Username,
		"exp":      jwt.TimeFunc().Add(7 * 24 * time.Hour).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, signErr := refreshToken.SignedString(s.secretKey)
	if signErr != nil {
		logger.Error("Failed to generate refresh token for new user", logger.Any("error", signErr))
		return nil, errs.NewUnexpectedError("Error generating refresh token: " + signErr.Error())
	}

	if err := s.repo.SaveRefreshToken(refreshTokenString); err != nil {
		logger.Error("Failed to save refresh token", logger.Any("error", err))
		return nil, err
	}

	return &dto.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshTokenString,
	}, nil
}

func (s *AuthService) Refresh(token string) (*dto.LoginResponse, *errs.AppError) {
	exists, err := s.repo.VerifyRefreshToken(token)
	if err != nil {
		return nil, err
	}
	if !exists {
		logger.Warn("Refresh token not found in database")
		return nil, errs.NewAuthenticationError("Invalid refresh token")
	}

	claims, tokenErr := utils.ExtractClaimsFromToken(token, func() []byte { return s.secretKey })
	if tokenErr != nil {
		logger.Error("Failed to parse refresh token", logger.Any("error", tokenErr))
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
	claims, tokenErr := utils.ExtractClaimsFromToken(token, func() []byte { return s.secretKey })
	if tokenErr != nil {
		logger.Error("Failed to parse token", logger.Any("error", tokenErr))
		return false, errs.NewAuthenticationError("Invalid token")
	}

	if exp, ok := claims["exp"].(float64); !ok || time.Unix(int64(exp), 0).Before(time.Now()) {
		logger.Warn("Token expired")
		return false, errs.NewAuthenticationError("Token expired")
	}

	role, ok := claims["role"].(string)
	if !ok {
		logger.Warn("Role not found in token")
		return false, errs.NewAuthenticationError("Invalid token format")
	}

	customerID, _ := claims["customer_id"].(string)

	isAuthorized := s.repo.VerifyPermission(role, customerID, routeName, vars)
	if !isAuthorized {
		logger.Warn("Permission denied",
			logger.String("role", role),
			logger.String("routeName", routeName))
		return false, errs.NewAuthenticationError("Unauthorized")
	}

	return true, nil
}

func (s *AuthService) GetSecretKey() []byte {
	return s.secretKey
}

func (s *AuthService) GetRolePermissions() domain.RolePermissions {
	return *domain.GetRolePermissions()
}
