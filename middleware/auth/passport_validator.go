// Package auth 提供认证中间件和工具函数
package auth

import (
	"context"
	"fmt"
	"time"

	passportv1 "passport-service/api/passport/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// PassportTokenValidator PassportService Token 验证器
// 所有服务都调用同一个 passport-service，所以直接在公共库中实现
// 服务间调用使用 gRPC（性能更好、类型安全）
// 注意：不需要接口抽象，因为所有服务都统一使用 passport-service
type PassportTokenValidator struct {
	client passportv1.PassportClient
	log    *log.Helper
}

// NewPassportTokenValidator 创建 PassportTokenValidator（gRPC 版本，推荐）
// grpcAddr: passport-service 的 gRPC 地址，如 "localhost:9100"
// timeout: gRPC 连接超时时间
func NewPassportTokenValidator(grpcAddr string, timeout time.Duration, logger log.Logger) (*PassportTokenValidator, error) {
	// 创建 gRPC 连接
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(grpcAddr),
		grpc.WithTimeout(timeout),
		grpc.WithMiddleware(
			recovery.Recovery(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial passport-service: %w", err)
	}

	// 创建 gRPC 客户端
	client := passportv1.NewPassportClient(conn)

	return &PassportTokenValidator{
		client: client,
		log:    log.NewHelper(logger),
	}, nil
}

// NewPassportTokenValidatorWithDefaults 使用默认超时时间创建
// 默认超时时间为 10 秒
func NewPassportTokenValidatorWithDefaults(grpcAddr string, logger log.Logger) (*PassportTokenValidator, error) {
	return NewPassportTokenValidator(grpcAddr, 10*time.Second, logger)
}

// ValidateToken 验证 token
func (v *PassportTokenValidator) ValidateToken(ctx context.Context, token string) (*UserClaims, error) {
	req := &passportv1.ValidateTokenRequest{
		Token: token,
	}

	resp, err := v.client.ValidateToken(ctx, req)
	if err != nil {
		v.log.Errorf("failed to validate token: %v", err)
		return nil, fmt.Errorf("validate token failed: %w", err)
	}

	return &UserClaims{
		UserID: resp.UserId,
		Role:   resp.Role,
	}, nil
}
