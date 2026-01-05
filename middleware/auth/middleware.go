// Package auth 提供认证中间件和工具函数
package auth

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Middleware 认证中间件，验证 token 并提取用户信息
// 如果 token 验证失败，不阻止请求，但不在 context 中设置用户信息
// 这样可以让某些接口允许匿名访问
// config: 认证配置，包含路由白名单
func Middleware(validator *PassportTokenValidator, config *Config, logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 检查路径是否在白名单中
			tr, ok := transport.FromServerContext(ctx)
			if ok {
				operation := tr.Operation()
				if config != nil && config.ShouldSkipPath(operation) {
					// 白名单路径，直接跳过认证
					return handler(ctx, req)
				}
			}

			// 从 transport 中获取 token
			if ok {
				var authHeader string

				// 尝试从 HTTP header 中获取 token
				if httpTr, ok := tr.(interface {
					RequestHeader() map[string][]string
				}); ok {
					headers := httpTr.RequestHeader()
					if authHeaders, ok := headers["Authorization"]; ok && len(authHeaders) > 0 {
						authHeader = authHeaders[0]
					}
					// 如果没有 Authorization header，尝试从 Cookie 中获取
					if authHeader == "" {
						if cookies, ok := headers["Cookie"]; ok && len(cookies) > 0 {
							// 解析 Cookie header
							for _, cookieStr := range cookies {
								parts := strings.Split(cookieStr, ";")
								for _, part := range parts {
									part = strings.TrimSpace(part)
									if strings.HasPrefix(part, "access_token=") {
										tokenValue := strings.TrimPrefix(part, "access_token=")
										authHeader = "Bearer " + tokenValue
										break
									}
								}
								if authHeader != "" {
									break
								}
							}
						}
					}
				}

				if authHeader != "" {
					// 解析 Bearer token
					parts := strings.Split(authHeader, " ")
					if len(parts) == 2 && parts[0] == "Bearer" {
						token := parts[1]

						// 验证 token
						claims, err := validator.ValidateToken(ctx, token)
						if err != nil {
							log.NewHelper(logger).Warnf("Token validation failed: %v", err)
							// 如果 token 验证失败，不阻止请求，但不在 context 中设置用户信息
							// 这样可以让某些接口允许匿名访问
						} else {
							// 将用户信息存储到 context 中
							ctx = WithUserClaims(ctx, claims)
						}
					}
				}
			}

			return handler(ctx, req)
		}
	}
}

// RequireAuth 要求认证的中间件，如果未认证则返回错误
func RequireAuth(validator *PassportTokenValidator, config *Config, logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 检查路径是否在白名单中
			tr, ok := transport.FromServerContext(ctx)
			if ok {
				operation := tr.Operation()
				if config != nil && config.ShouldSkipPath(operation) {
					// 白名单路径，直接跳过认证
					return handler(ctx, req)
				}
			}

			claims, ok := GetUserClaimsFromContext(ctx)
			if !ok || claims.UserID == "" {
				return nil, status.Error(codes.Unauthenticated, "authentication required")
			}

			return handler(ctx, req)
		}
	}
}

// RequireRole 要求特定角色的中间件
func RequireRole(requiredRole string, validator *PassportTokenValidator, config *Config, logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 检查路径是否在白名单中
			tr, ok := transport.FromServerContext(ctx)
			if ok {
				operation := tr.Operation()
				if config != nil && config.ShouldSkipPath(operation) {
					// 白名单路径，直接跳过认证
					return handler(ctx, req)
				}
			}

			claims, ok := GetUserClaimsFromContext(ctx)
			if !ok {
				return nil, status.Error(codes.Unauthenticated, "authentication required")
			}

			if claims.Role != requiredRole {
				return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
			}

			return handler(ctx, req)
		}
	}
}



