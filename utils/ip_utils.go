// Package utils 提供通用工具函数
package utils

import (
	"context"
	"fmt"
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
// 排除私有IP、本地IP、未指定地址和组播地址
// 使用 net.ParseIP 和 net.IP 的标准方法进行完整的 IP 地址验证
func IsValidPublicIP(ip string) bool {
	if ip == "" {
		return false
	}

	// 解析 IP 地址
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		// 如果无法解析，可能是无效的 IP 地址，返回 false
		return false
	}

	// 排除本地回环地址
	if parsedIP.IsLoopback() {
		return false
	}

	// 排除链路本地地址（IPv4 和 IPv6）
	if parsedIP.IsLinkLocalUnicast() || parsedIP.IsLinkLocalMulticast() {
		return false
	}

	// 排除私有网络地址
	if parsedIP.IsPrivate() {
		return false
	}

	// 排除未指定地址（0.0.0.0 和 ::）
	if parsedIP.IsUnspecified() {
		return false
	}

	// 排除组播地址
	if parsedIP.IsMulticast() {
		return false
	}

	// 如果通过了所有检查，认为是有效的公网 IP
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

// ExtractFirstPublicIP 从字符串中提取第一个公网 IP。
//
// 典型场景：解析 X-Forwarded-For / X-Real-IP 等头部值。
// 支持：
// - 多个 IP 以逗号分隔（取第一个公网 IP）
// - IPv6 方括号格式（如 [2001:db8::1]）
// 返回约定：
// - 若存在公网 IP，返回第一个公网 IP
// - 若不存在公网 IP，但能解析出任意合法 IP，则返回第一个合法 IP
// - 若无法解析任何 IP，则返回空字符串
func ExtractFirstPublicIP(ipStr string) string {
	ipStr = strings.TrimSpace(ipStr)
	if ipStr == "" {
		return ""
	}
	parts := strings.Split(ipStr, ",")
	for _, part := range parts {
		p := strings.TrimSpace(part)
		p = strings.Trim(p, "[]")
		if net.ParseIP(p) == nil {
			continue
		}
		if IsValidPublicIP(p) {
			return p
		}
	}
	for _, part := range parts {
		p := strings.TrimSpace(part)
		p = strings.Trim(p, "[]")
		if net.ParseIP(p) != nil {
			return p
		}
	}
	return ""
}

// MaskIP 对 IP 做脱敏展示。
//
// - IPv4：保留前两段（如 1.2.*.*）
// - IPv6：保留前 8 个字符前缀（如 2001:db8...）
// 解析失败时返回空字符串。
func MaskIP(ipStr string) string {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		return ""
	}
	if v4 := ip.To4(); v4 != nil {
		return fmt.Sprintf("%d.%d.*.*", v4[0], v4[1])
	}
	s := ip.String()
	if len(s) <= 8 {
		return s
	}
	return s[:8] + "..."
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
