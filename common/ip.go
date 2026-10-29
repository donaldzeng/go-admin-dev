package common

import (
	"github.com/gin-gonic/gin"
)

// TrustedProxies 是允许改写客户端 IP 头（X-Forwarded-For / X-Real-IP）的
// 网段，即部署链路上的自有反向代理：NAS nginx 与 Go api 之间是 docker
// bridge 私有网段。在 server 启动时通过 gin 的 SetTrustedProxies 安装。
//
// 2026-10-29 入侵事件加固的一部分：此前 GetClientIP 无条件信任请求头，
// 客户端随手伪造一个 X-Forwarded-For 就能伪装成任意来源——登录日志记录
// 假 IP，任何按 IP 封锁的防线（防爆破守卫、nginx deny）都会被绕过。
var TrustedProxies = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
}

// GetClientIP 返回发起请求的客户端真实 IP。
//
// 解析交给 gin 的 c.ClientIP()：它只在直连方（RemoteIP）属于可信代理时
// 才读转发头，并且从 X-Forwarded-For 右端向左取第一个非可信地址——攻击
// 者自己塞在头最左端的伪造条目永远不会被取到，最右端由云上 nginx 追加的
// 真实来源无法伪造。直连（本机调试、内网直打 10086 端口）时直接返回
// RemoteIP，转发头整个不被理会。
func GetClientIP(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "" {
		ip = c.RemoteIP()
	}
	if ip == "" {
		// 兜底为本地地址
		ip = "127.0.0.1"
	}
	return ip
}
