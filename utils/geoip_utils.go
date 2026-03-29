package utils

import (
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

// GeoIPResolver 基于 MaxMind GeoIP2 mmdb 的离线 IP 地理位置解析器。
//
// 设计要点：
// - 延迟加载：首次查询时才打开 mmdb 文件，减少无效 IO
// - 并发安全：通过 sync.Once 保证只初始化一次
// - 最小能力：只提供城市级别解析，避免把业务细节耦合进公共库
type GeoIPResolver struct {
	dbPath string
	once   sync.Once
	reader *geoip2.Reader
	err    error
}

// NewGeoIPResolver 创建 GeoIPResolver。
// dbPath 为空时不会报错，但 LookupCity 会返回 ok=false。
func NewGeoIPResolver(dbPath string) *GeoIPResolver {
	return &GeoIPResolver{dbPath: dbPath}
}

// LookupCity 根据 IP 查询国家/省/市。
//
// 返回约定：
// - ok=false：表示解析不可用或没有有效结果（如 dbPath 为空、mmdb 打开失败、IP 非法等）
// - 国家/省/市优先返回中文（zh-CN），若不存在则回退英文（en）
func (r *GeoIPResolver) LookupCity(ipStr string) (country, province, city string, ok bool) {
	if r == nil || r.dbPath == "" {
		return "", "", "", false
	}
	r.once.Do(func() {
		r.reader, r.err = geoip2.Open(r.dbPath)
	})
	if r.err != nil || r.reader == nil {
		return "", "", "", false
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", "", "", false
	}

	rec, err := r.reader.City(ip)
	if err != nil {
		return "", "", "", false
	}

	country = rec.Country.Names["zh-CN"]
	if country == "" {
		country = rec.Country.Names["en"]
	}
	if len(rec.Subdivisions) > 0 {
		province = rec.Subdivisions[0].Names["zh-CN"]
		if province == "" {
			province = rec.Subdivisions[0].Names["en"]
		}
	}
	city = rec.City.Names["zh-CN"]
	if city == "" {
		city = rec.City.Names["en"]
	}

	if country == "" && province == "" && city == "" {
		return "", "", "", false
	}
	return country, province, city, true
}

// Close 关闭底层 mmdb reader。
// 说明：GeoIPResolver 设计为进程级复用，通常不必显式调用；仅在需要释放文件句柄时使用。
func (r *GeoIPResolver) Close() error {
	if r == nil || r.reader == nil {
		return nil
	}
	return r.reader.Close()
}
