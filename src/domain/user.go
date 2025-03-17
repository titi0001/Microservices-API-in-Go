package domain

type UserDetails struct {
	Username string
	UserID   string
	
	
	Role      string   
	Roles     []string 
	
	CustomClaims map[string]interface{}
}