// Package auth 提供认证中间件和工具函数
package auth

import "context"

// userClaimsKey 是 context 中存储用户声明信息的键
type userClaimsKey struct{}

// UserClaimsKey 导出用户声明键，供外部使用
var UserClaimsKey = userClaimsKey{}

// UserClaims 用户声明信息
type UserClaims struct {
	UserID string
	Role   string
}

// WithUserClaims 将用户声明信息存入 context
func WithUserClaims(ctx context.Context, claims *UserClaims) context.Context {
	return context.WithValue(ctx, UserClaimsKey, claims)
}

// GetUserClaimsFromContext 从 context 中获取用户声明信息
func GetUserClaimsFromContext(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*UserClaims)
	return claims, ok
}

// GetUserIDFromContext 从 context 中获取用户ID（便捷方法）
func GetUserIDFromContext(ctx context.Context) string {
	claims, ok := GetUserClaimsFromContext(ctx)
	if !ok {
		return ""
	}
	return claims.UserID
}

// GetUserRoleFromContext 从 context 中获取用户角色（便捷方法）
func GetUserRoleFromContext(ctx context.Context) string {
	claims, ok := GetUserClaimsFromContext(ctx)
	if !ok {
		return ""
	}
	return claims.Role
}





