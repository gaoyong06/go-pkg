// Package utils 提供通用工具函数
package utils

import (
	"context"
	"net"
	"strings"

	"github.com/go-kratos/kratos/v2/transport"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// GetUserAgent 从 context 中获取 User-Agent
// 使用 kratos transport 接口，适用于所有使用 kratos 框架的项目
func GetUserAgent(ctx context.Context) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return ""
	}
	return tr.RequestHeader().Get("User-Agent")
}

// GetClientIPRaw 从 context 中获取客户端IP（不验证是否为公网IP）
// 用于审计日志等需要记录所有IP的场景
func GetClientIPRaw(ctx context.Context) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return ""
	}
	header := tr.RequestHeader()
	
	// 1. 优先使用 X-Forwarded-For（代理服务器转发）
	if xff := header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For 可能包含多个 IP，取第一个
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 2. 使用 X-Real-IP（Nginx 等代理服务器）
	if xri := header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// 3. 使用 RemoteAddr（直连IP）
	if ht, ok := tr.(*kratoshttp.Transport); ok {
		remoteAddr := ht.Request().RemoteAddr
		if remoteAddr != "" {
			ip, _, err := net.SplitHostPort(remoteAddr)
			if err == nil {
				return ip
			}
			// 如果 SplitHostPort 失败，可能 RemoteAddr 就是 IP 地址
			return remoteAddr
		}
	}

	return ""
}

// EnrichRequestInfo 从 HTTP 请求中提取 IP 和 UserAgent，并添加到 context
// 用于审计日志记录等场景
// 注意：使用 GetClientIPRaw 而不是 GetClientIP，因为审计日志需要记录所有IP（包括私有IP）
func EnrichRequestInfo(ctx context.Context) context.Context {
	// 提取 IP 地址（不验证是否为公网IP，记录所有IP）
	if ip := GetClientIPRaw(ctx); ip != "" {
		ctx = context.WithValue(ctx, "ip_address", ip)
	}

	// 提取 UserAgent
	if userAgent := GetUserAgent(ctx); userAgent != "" {
		ctx = context.WithValue(ctx, "user_agent", userAgent)
	}

	return ctx
}

