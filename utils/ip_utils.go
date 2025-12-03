// Package utils 提供通用工具函数
package utils

import (
	"context"
	"net"
	"strings"

	"github.com/go-kratos/kratos/v2/transport"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// GetClientIP 从 context 中获取客户端IP
// 使用 kratos transport 接口，适用于所有使用 kratos 框架的项目
func GetClientIP(ctx context.Context) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return ""
	}
	return getClientIPFromTransport(tr)
}

// IsValidPublicIP 验证IP地址是否为有效的公网IP
// 排除私有IP和本地IP
func IsValidPublicIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 排除私有IP和本地IP
	if parsedIP.IsLoopback() || parsedIP.IsLinkLocalUnicast() || parsedIP.IsLinkLocalMulticast() {
		return false
	}

	// 排除私有网络地址
	if parsedIP.IsPrivate() {
		return false
	}

	return true
}

// IsPrivateIP 检查是否为私有IP
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.IsPrivate()
}

// =========================================== private functions ===========================================

// getClientIPFromTransport 从 kratos transport 中获取客户端IP（内部方法）
func getClientIPFromTransport(tr transport.Transporter) string {
	// 1. 检查 X-Forwarded-For 头（代理服务器设置）
	if xff := tr.RequestHeader().Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For 可能包含多个IP，取第一个
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if IsValidPublicIP(ip) {
				return ip
			}
		}
	}

	// 2. 检查 X-Real-IP 头（Nginx等设置）
	if xri := tr.RequestHeader().Get("X-Real-IP"); xri != "" {
		if IsValidPublicIP(xri) {
			return xri
		}
	}

	// 3. 检查 X-Forwarded 头
	if xf := tr.RequestHeader().Get("X-Forwarded"); xf != "" {
		if IsValidPublicIP(xf) {
			return xf
		}
	}

	// 4. 检查 CF-Connecting-IP 头（Cloudflare）
	if cfip := tr.RequestHeader().Get("CF-Connecting-IP"); cfip != "" {
		if IsValidPublicIP(cfip) {
			return cfip
		}
	}

	// 5. 使用 RemoteAddr（直连IP）
	var remoteAddr string
	if ht, ok := tr.(*kratoshttp.Transport); ok {
		remoteAddr = ht.Request().RemoteAddr
	}
	if remoteAddr != "" {
		ip, _, err := net.SplitHostPort(remoteAddr)
		if err == nil && IsValidPublicIP(ip) {
			return ip
		}
		// 如果 SplitHostPort 失败，可能 RemoteAddr 就是 IP 地址
		if IsValidPublicIP(remoteAddr) {
			return remoteAddr
		}
	}

	// 6. 如果都获取不到，返回 RemoteAddr 的原始值
	return remoteAddr
}
