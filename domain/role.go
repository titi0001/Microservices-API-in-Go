package domain

import (
	"strings"
	"sync"

	"github.com/titi0001/Microservices-API-in-Go/logger"
)

var (
	singletonRolePermissions *RolePermissions
	once                     sync.Once
)

// RolePermissions gerencia permissões por papel
type RolePermissions struct {
	rolePermissions map[string][]string
}

// GetRolePermissions retorna a instância singleton de permissões
func GetRolePermissions() *RolePermissions {
	once.Do(func() {
		permissions := map[string][]string{
			"admin": {"GetAllCustomers", "GetCustomer", "NewAccount", "NewTransaction", "GetRolePermissions"},
			"user":  {"GetCustomer", "NewTransaction"},
		}
		singletonRolePermissions = &RolePermissions{rolePermissions: permissions}
		logger.Info("Initialized RolePermissions singleton",
			logger.Int("role_count", len(permissions)))
	})
	return singletonRolePermissions
}

// IsAuthorizedFor verifica se um papel tem permissão para uma rota
func (p *RolePermissions) IsAuthorizedFor(role, routeName string) bool {
	normalizedRole := strings.ToLower(strings.TrimSpace(role))
	normalizedRouteName := strings.TrimSpace(routeName)

	if len(p.rolePermissions) == 0 {
		logger.Warn("RolePermissions map is empty, using default permissions")
		p.rolePermissions = defaultPermissions()
	}

	perms, exists := p.rolePermissions[normalizedRole]
	if !exists {
		logger.Warn("Role not found", logger.String("role", normalizedRole))
		if normalizedRole == "admin" {
			logger.Info("Emergency bypass for admin role")
			return true
		}
		return false
	}

	for _, perm := range perms {
		if perm == normalizedRouteName {
			return true
		}
	}
	logger.Warn("Permission not found",
		logger.String("role", normalizedRole),
		logger.String("routeName", normalizedRouteName))
	return false
}

// GetAllPermissions retorna todas as permissões únicas
func (p *RolePermissions) GetAllPermissions() []string {
	if len(p.rolePermissions) == 0 {
		logger.Warn("RolePermissions map is empty, using default permissions")
		p.rolePermissions = defaultPermissions()
	}

	uniquePerms := make(map[string]struct{})
	for _, perms := range p.rolePermissions {
		for _, perm := range perms {
			uniquePerms[perm] = struct{}{}
		}
	}

	result := make([]string, 0, len(uniquePerms))
	for perm := range uniquePerms {
		result = append(result, perm)
	}
	return result
}

// defaultPermissions retorna as permissões padrão
func defaultPermissions() map[string][]string {
	return map[string][]string{
		"admin": {"GetAllCustomers", "GetCustomer", "NewAccount", "NewTransaction", "GetRolePermissions"},
		"user":  {"GetCustomer", "NewTransaction"},
	}
}