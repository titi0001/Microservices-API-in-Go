package dto

type UserDetails struct {
	Username     string                 `json:"username"`
	UserID       string                 `json:"user_id"`
	Role         string                 `json:"role"`
	Roles        []string               `json:"roles"`
	CustomClaims map[string]interface{} `json:"custom_claims"`
}