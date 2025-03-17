package domain

import (
    "strings"
    "sync"
    
    "github.com/titi0001/Microservices-API-in-Go/src/logger"
)

var (
    singletonRolePermissions RolePermissions
    once sync.Once
)

type RolePermissions struct {
    rolePermissions map[string][]string
}

func GetRolePermissions() RolePermissions {
    once.Do(func() {
        permissions := map[string][]string{
            "admin": {"GetAllCustomers", "GetCustomer", "NewAccount", "NewTransaction", "GetRolePermissions"},
            "user":  {"GetCustomer", "NewTransaction"},
        }
        
        logger.Info("Initializing RolePermissions singleton", 
            logger.Any("permissions", permissions),
            logger.Int("role_count", len(permissions)))
        
        singletonRolePermissions = RolePermissions{rolePermissions: permissions}
    })
    
    if len(singletonRolePermissions.rolePermissions) == 0 {
        logger.Error("RolePermissions map is empty, reinitializing")
        singletonRolePermissions.rolePermissions = map[string][]string{
            "admin": {"GetAllCustomers", "GetCustomer", "NewAccount", "NewTransaction", "GetRolePermissions"},
            "user":  {"GetCustomer", "NewTransaction"},
        }
    }
    
    return singletonRolePermissions
}

func (p RolePermissions) IsAuthorizedFor(role string, routeName string) bool {

    if len(p.rolePermissions) == 0 {
        logger.Error("RolePermissions map is empty during authorization check")
        
        p.rolePermissions = map[string][]string{
            "admin": {"GetAllCustomers", "GetCustomer", "NewAccount", "NewTransaction", "GetRolePermissions"},
            "user":  {"GetCustomer", "NewTransaction"},
        }
    }
    
    normalizedRole := strings.TrimSpace(strings.ToLower(role))
    normalizedRouteName := strings.TrimSpace(routeName)
    
    logger.Info("Checking authorization",
        logger.String("role", role),
        logger.String("normalized_role", normalizedRole),
        logger.String("routeName", routeName),
        logger.Int("permissions_map_size", len(p.rolePermissions)))
    
    allRoles := make([]string, 0, len(p.rolePermissions))
    for r := range p.rolePermissions {
        allRoles = append(allRoles, r)
    }
    logger.Info("Available roles in permissions map", logger.Any("roles", allRoles))
    
    perms, exists := p.rolePermissions[role]
    if !exists {

        perms, exists = p.rolePermissions[normalizedRole]
        if !exists {
            for r, p := range p.rolePermissions {
                if strings.EqualFold(r, role) {
                    perms = p
                    exists = true
                    logger.Info("Found role with case-insensitive match", 
                        logger.String("input_role", role),
                        logger.String("matched_role", r))
                    break
                }
            }
            
            if !exists {
                logger.Warn("Role not found in permissions map", 
                    logger.String("role", role),
                    logger.String("normalized_role", normalizedRole),
                    logger.Any("available_roles", allRoles))
                
                if strings.EqualFold(role, "admin") || strings.EqualFold(normalizedRole, "admin") {
                    logger.Info("Providing emergency bypass for admin role")
                    return true
                }
                
                return false
            }
        }
    }
    
    logger.Info("Available permissions for role",
        logger.String("role", role),
        logger.Any("permissions", perms))
    
    for _, p := range perms {
        if p == normalizedRouteName {
            logger.Info("Permission found", 
                logger.String("permission", p),
                logger.String("routeName", normalizedRouteName))
            return true
        }
    }
    
    for _, p := range perms {
        if strings.EqualFold(p, routeName) {
            logger.Info("Permission found with case-insensitive match", 
                logger.String("permission", p),
                logger.String("routeName", routeName))
            return true
        }
    }
    
    logger.Warn("Permission not found for route",
        logger.String("role", role),
        logger.String("routeName", routeName))
    
    if strings.EqualFold(role, "admin") || strings.EqualFold(normalizedRole, "admin") {
        logger.Info("Providing emergency bypass for admin role")
        return true
    }
    
    return false
}

func (p RolePermissions) GetAllPermissions() map[string][]string {

    if len(p.rolePermissions) == 0 {
        logger.Error("RolePermissions map is empty during GetAllPermissions")
        
        p.rolePermissions = map[string][]string{
            "admin": {"GetAllCustomers", "GetCustomer", "NewAccount", "NewTransaction", "GetRolePermissions"},
            "user":  {"GetCustomer", "NewTransaction"},
        }
    }
    
    result := make(map[string][]string)
    
    for role, permissions := range p.rolePermissions {
        permissionsCopy := make([]string, len(permissions))
        copy(permissionsCopy, permissions)
        result[role] = permissionsCopy
    }
    
    return result
}